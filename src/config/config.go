package config

import (
	"flag"
	"os"
	"path"

	"github.com/Unknwon/goconfig"
	log "github.com/cihub/seelog"
	"github.com/lyokalita/naspublic.ftserver/src/utils"
)

var (
	ServerHost          string
	ServerPort          int
	ApiPath             string
	PublicDirectoryRoot string
	NumCore             int
	WebfrontendOrigin   []string
	JwtSecret           []byte = []byte("123")
	SignSecret          []byte = []byte(utils.GetRandomBytes(32))
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

	err = os.MkdirAll(PublicDirectoryRoot, os.ModePerm)
	if err != nil {
		log.Errorf("failed to create download directory %s, err: ", PublicDirectoryRoot, err)
	}
	log.Debugf("successfully loaded config, NumCore: %d, Cors: %v", NumCore, WebfrontendOrigin)
}
