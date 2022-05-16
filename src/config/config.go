package config

import (
	"flag"
	"os"
	"path"

	"github.com/Unknwon/goconfig"
	log "github.com/cihub/seelog"
)

var (
	ServerHost          string
	ServerPort          int
	ApiPath             string
	PublicDirectoryRoot string
	NumCore             int
	WebfrontendOrigin   []string
	AuthOrigin          []string
	JwtSecret           []byte
	SignSecret          []byte
	AuthSecret          string
	SSLCertPath         string
	SSLKeyPath          string
)

var (
	configPath       string
	seelogConfigPath string
)

func Init() {
	ParseParam()
	SetupLogger()
	LoadConfig()
}

func ParseParam() {
	conf := flag.String("config", "./conf/config.ini", "config path")
	log := flag.String("seelog", "./conf/seelog.xml", "seelog config path")
	flag.Parse()

	configPath = *conf
	seelogConfigPath = *log
}

func SetupLogger() {
	logger, err := log.LoggerFromConfigAsFile(seelogConfigPath)
	if err != nil {
		logger, _ = log.LoggerFromConfigAsString(`<seelog minlevel="debug"><outputs formatid="main"><buffered size="10000" flushperiod="1000"><rollingfile type="size" filename="log/ftserver.log" maxsize="6http.StatusBadRequest0000" maxrolls="50"/></buffered></outputs><formats><format id="main" format="[%Date(2006-01-02 15:04:05.999 PM MST)] [%Level] [%File:%FuncShort#%Line] %Msg%n"/></formats></seelog>`)

	}
	log.ReplaceLogger(logger)
	log.Debugf("successfully setup logger from %s", seelogConfigPath)
}

func LoadConfig() {
	cfg, err := goconfig.LoadConfigFile(configPath)
	if err != nil {
		log.Errorf("failed to open config file %s, err: %v", configPath, err)
		panic(err)
	}

	ServerHost = cfg.MustValue("server", "host", "")
	ServerPort = cfg.MustInt("server", "port", 4500)
	ApiPath = cfg.MustValue("server", "path", "/api/nas/v0")
	PublicDirectoryRoot = cfg.MustValue("directory_root", "path", "./temp/")
	PublicDirectoryRoot = path.Join(PublicDirectoryRoot)
	NumCore = cfg.MustInt("hardware", "num_core", 4)
	WebfrontendOrigin = cfg.MustValueArray("cors", "webfrontend", ",")
	AuthOrigin = cfg.MustValueArray("cors", "auth", ",")
	SSLCertPath = cfg.MustValue("ssl", "cert", ".cert/localhost.cert")
	SSLCertPath = cfg.MustValue("ssl", "key", ".cert/localhost.key")

	err = os.MkdirAll(PublicDirectoryRoot, os.ModePerm)
	if err != nil {
		log.Errorf("failed to create download directory %s, err: ", PublicDirectoryRoot, err)
	}

	JwtSecret = []byte(os.Getenv("NASPUBLIC_JWT_SECRET"))
	SignSecret = []byte(os.Getenv("NASPUBLIC_SIGN_SECRET"))
	AuthSecret = os.Getenv("NASPUBLIC_AUTH_SECRET")
	if len(JwtSecret) == 0 || len(SignSecret) == 0 || AuthSecret == "" {
		panic("failed to load secrets")
	}
	log.Debugf("successfully loaded config, public root: %s, NumCore: %d, Cors: frontend: %v, auth: %v, ssl cert path: %s, ssl key path: %s", PublicDirectoryRoot, NumCore, WebfrontendOrigin, AuthOrigin, SSLCertPath, SSLKeyPath)
}
