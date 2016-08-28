package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"text/template"

	yaml "gopkg.in/yaml.v2"
)

var (
	TemplateDir         = "./tmpl"
	TemplateComposeEtcd = TemplateDir + "/compose-etcd.yaml"
	Default             = Spec{
		Image:            "quay.io/coreos/etcd:latest",
		ClientPort:       "2379",
		PeerPort:         "2380",
		ListenPublicAddr: "0.0.0.0",
		ListenClientAddr: "0.0.0.0",
		ListenPeerAddr:   "0.0.0.0",
		ClusterState:     "new",
		Debug:            "false",
	}
	ComposeDir = "./compose"
)

type Config struct {
	Template Spec   `yaml:"template"`
	Specs    []Spec `yaml:"spec"`
}

type Spec struct {
	Name             string `yaml:"name"`
	Domain           string `yaml:"domain"`
	Image            string `yaml:"image"`
	ClientPort       string `yaml:"client_port"`
	PeerPort         string `yaml:"peer_port"`
	ListenPublicAddr string `yaml:"listen_public_addr"`
	ListenClientAddr string `yaml:"listen_client_addr"`
	ListenPeerAddr   string `yaml:"listen_peer_addr"`
	ClusterState     string `yaml:"cluster_state"`
	Token            string `yaml:"token"`
	Debug            string `yaml:"debug"`

	FQDN               string
	ListenClientURL    string
	AdvertiseClientURL string
	ListenPeerURL      string
	AdvertisePeerURL   string
	InitialCluster     []string
}

func (sp *Spec) inherit(src Spec) {
	if len(sp.Domain) == 0 {
		sp.Domain = src.Domain
	}

	if len(sp.Image) == 0 {
		sp.Image = src.Image
	}

	if len(sp.ClientPort) == 0 {
		sp.ClientPort = src.ClientPort
	}

	if len(sp.PeerPort) == 0 {
		sp.PeerPort = src.PeerPort
	}

	if len(sp.ListenPublicAddr) == 0 {
		sp.ListenPublicAddr = src.ListenPublicAddr
	}

	if len(sp.ListenClientAddr) == 0 {
		sp.ListenClientAddr = src.ListenClientAddr
	}

	if len(sp.ListenPeerAddr) == 0 {
		sp.ListenPeerAddr = src.ListenPeerAddr
	}

	if len(sp.ClusterState) == 0 {
		sp.ClusterState = src.ClusterState
	}

	if len(sp.Token) == 0 {
		sp.Token = src.Token
	}

	if len(sp.Debug) == 0 {
		sp.Debug = src.Debug
	}
}

func (sp *Spec) complete() {

	if len(sp.Domain) == 0 {
		sp.FQDN = sp.Name
	} else if sp.Domain[:1] == "." {
		sp.FQDN = sp.Name + sp.Domain
	} else {
		sp.FQDN = sp.Name + "." + sp.Domain
	}

	sp.ListenClientURL = "https://" + sp.ListenClientAddr + ":" + sp.ClientPort
	sp.AdvertiseClientURL = "https://" + sp.FQDN + ":" + sp.ClientPort
	sp.ListenPeerURL = "https://" + sp.ListenPeerAddr + ":" + sp.PeerPort
	sp.AdvertisePeerURL = "https://" + sp.FQDN + ":" + sp.PeerPort
}

func (sp *Spec) validate(i int) {
	if len(sp.Name) == 0 {
		log.Fatalf("'name' must be specified for %d-th spec", i)
	}
}

func (sp *Spec) generate(tmpl *template.Template) {
	outDir := ComposeDir + "/" + sp.Name

	err := os.MkdirAll(outDir, 0700)
	if err != nil {
		log.Fatalf("MkdirAll(%s): %s", outDir, err)
	}

	outPath := outDir + "/docker-compose.yaml"
	out, err := os.OpenFile(outPath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Open(%s): %s", outPath, err)
	}
	defer out.Close()

	err = tmpl.Execute(out, sp)
	if err != nil {
		log.Fatalf("template.Execute(): %s", err)
	}
}

func join(ss []string) string {
	return strings.Join(ss, ",")
}

func loadTemplate() *template.Template {
	f, err := os.Open(TemplateComposeEtcd)
	if err != nil {
		log.Fatalf("%s: %s", TemplateComposeEtcd, err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatalf("ReadAll(%s): %s", TemplateComposeEtcd, err)
	}

	tmplsrc := string(b)

	tmpl, err := template.New("docker-compose.yaml").
		Funcs(template.FuncMap{"join": join}).
		Parse(tmplsrc)
	if err != nil {
		log.Fatalf("template.New: %s", err)
	}

	return tmpl
}

func readConfig(w io.Reader) []Spec {
	var cfg Config

	b, err := ioutil.ReadAll(w)
	if err != nil {
		log.Fatalf("ReadAll(stdin): %s", err)
	}

	err = yaml.Unmarshal(b, &cfg)
	if err != nil {
		log.Fatalf("yaml.Unmarshal(): %s", err)
	}

	cfg.Template.inherit(Default)
	if len(cfg.Template.Name) != 0 {
		log.Fatalf("Cannot be specified 'name' field in 'template'")
	}

	peerURLs := make([]string, len(cfg.Specs))
	for i, _ := range cfg.Specs {
		sp := &cfg.Specs[i]
		sp.inherit(cfg.Template)
		sp.complete()
		sp.validate(i)
		sp.InitialCluster = peerURLs
		peerURLs[i] = sp.Name + "=" + sp.AdvertisePeerURL
	}

	return cfg.Specs
}

func main() {
	specs := readConfig(os.Stdin)
	tmpl := loadTemplate()

	for _, sp := range specs {
		sp.generate(tmpl)
	}

}
