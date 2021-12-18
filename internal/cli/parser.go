package cli

import (
	"errors"
	"fofax/internal/printer"
	"fofax/internal/utils"
	"github.com/projectdiscovery/goflags"
	"os"
	"path/filepath"
)

const (
	NONE_Mode = iota
	Stdin_Mode
	Query_Mode
	File_Mode
)

type Options struct {
	Mode int
	query
	queryOfFile
	filter
	config
	Version bool
	Use     bool
	// 标准输入
	Stdin bool
}

type query struct {
	// 查询的语句
	Query string `key:"-q"`
	// 计算 ico 文件
	IconFilePath string `key:"-if"`
	// 单个 URL，计算 icon hash 后进行查询
	UrlIcon string `key:"-ui"`
	// 输入 url(https) 查询其证书
	PeerCertificates string `key:"-uc"`
}

type queryOfFile struct {
	// 从文件中进行查询
	QueryFile string `key:"-qf"`
	// 批量 URL，计算 icon hash 后进行查询
	UrlIconFile string `key:"-iuf"`
	// 通过文件中多个的多个 url 查询其证书
	PeerCertificatesFile string `key:"-ucf"`
}

type filter struct {
	// 填写需要的另一个字段如，port
	FetchOneField string
	// 提取指定根域名的所有 title
	FetchTitlesOfDomain bool
	// 提取完整的 hostInfo，带有 protocol
	FetchFullHostInfo bool
	// 排除干扰
	Exclude bool
	// 排除国家
	ExcludeCountryCN bool
	// 去重 ,好像没用？
	//UniqByIP bool
	// 搜索数
	FetchSize int
}
type config struct {
	// fofa 地址
	FoFaURL   string
	FoFaEmail string
	FoFaKey   string
	// 脱敏密码
	FoFaKeyFake string
	Proxy       string
	Debug       bool
	ConfigFile  string
}

var (
	args  *Options
	flags *goflags.FlagSet
)

func initOptions() {
	args = new(Options)
	args.FoFaEmail = os.Getenv("FOFA_EMAIL")
	args.FoFaKey = os.Getenv("FOFA_KEY")
	args.FoFaURL = "https://fofa.so"
	args.FetchSize = 100
}

func init() {
	initOptions()
	flags = goflags.NewFlagSet()
	flags.SetDescription("FoFax is a command line fofa query tool, simple is the best!")
	//flags.StringVar(&args.ConfigFile, "config", args.ConfigFile, "fofadump configuration file.The file reading order")
	createGroup(
		flags, "config", "配置项",
		flags.StringVarP(&args.FoFaEmail, "fofa-email", "email", args.FoFaEmail, "Fofa API Email"),
		flags.StringVarP(&args.FoFaKey, "fofakey", "key", args.FoFaKey, "Fofa API Key"),
		flags.StringVarP(&args.Proxy, "proxy", "p", "", "proxy for http like http://127.0.0.1:8080"),
		flags.StringVar(&args.FoFaURL, "fofa-url", args.FoFaURL, "Fofa url"),
		flags.BoolVar(&args.Debug, "debug", false, "开启 debug 模式"),
	)
	createGroup(
		flags, "filters", "过滤项",
		flags.IntVarP(&args.FetchSize, "fetch-size", "fs", args.FetchSize, "最大查询数,默认 100"),
		flags.BoolVarP(&args.Exclude, "exclude", "e", args.Exclude, "排除干扰"),
		flags.BoolVarP(&args.ExcludeCountryCN, "exclude-country", "ec", false, "过滤 CN"),
		// 好像没用
		//flags.BoolVarP(&args.UniqByIP, "unique-by-ip", "ubi", args.UniqByIP, "以IP的方式进行去重"),
		flags.BoolVarP(&args.FetchFullHostInfo, "fetch-fullHost-info", "ffi", args.FetchFullHostInfo, "提取完整的 hostinfo,带有 protocol"),
		flags.BoolVarP(&args.FetchTitlesOfDomain, "fetch-titles-ofDomain", "fto", args.FetchTitlesOfDomain, "提取指定根域名的 title"),
		//flags.StringVarP(&args.FetchOneField, "fetch-one-field", "fof", args.FetchOneField, "填写需要的另一个字段如，port"),
	)
	createGroup(
		flags, "query", "单个 Query/cert/icon 搜索项",
		flags.StringVarP(&args.Query, "query", "q", args.Query, "FoFa 查询语句"),
		flags.StringVarP(&args.PeerCertificates, "url-cert", "uc", args.PeerCertificates, "输入 url(https) 查询证书"),
		flags.StringVarP(&args.UrlIcon, "url-to-icon-hash", "ui", args.UrlIcon, "通过 URL，计算 icon hash 后进行查询"),
		flags.StringVarP(&args.IconFilePath, "icon-file-path", "if", args.IconFilePath, "通过 ico 文件，计算 icon hash 后进行查询"),
	)
	createGroup(
		flags, "queryFile", "多个 Query/cert/icon 搜索项",
		flags.StringVarP(&args.QueryFile, "query-file", "qf", args.QueryFile, "加载文件，查询多个语句"),
		flags.StringVarP(&args.PeerCertificatesFile, "url-cert-file", "ucf", args.UrlIconFile, "读取文件中的URL，计算 cert 后进行查询"),
		flags.StringVarP(&args.UrlIconFile, "icon-hash-url-file", "iuf", args.UrlIconFile, "读取文件中的URL，计算 icon hash 后进行查询"),
	)
	flags.BoolVarP(&args.Version, "version", "v", false, "Show version of fofadump")
	flags.BoolVar(&args.Use, "use", false, "Query syntax reference")
	err := flags.Parse()
	if err != nil {
		printer.Error(printer.HandlerLine("Parse err :" + err.Error()))
		os.Exit(1)
	}
}

func createGroup(flagSet *goflags.FlagSet, name, desc string, flags ...*goflags.FlagData) {
	flagSet.SetGroup(name, desc)
	for _, currentFlag := range flags {
		currentFlag.Group(name)
	}
}

func ParseOptions() *Options {

	args.Stdin = utils.HasStdin()
	if !args.Stdin {
		banner()
	} else {
		args.Mode = Stdin_Mode
	}

	if args.Version {
		printer.Infof("Version: %s", FoFaXVersion)
		printer.Infof("Branch: %s", Branch)
		printer.Infof("Commit: %s", Commit)
		printer.Infof("Date: %s", Date)
		os.Exit(0)
	}

	if args.Use {
		ShowUsage()
		os.Exit(0)
	}

	// 检查基本信息
	checkFoFaInfo()

	// 检查参数是否互斥
	err := checkMutFlags()
	if err != nil {
		printer.Error(printer.HandlerLine(err.Error()))
		os.Exit(1)
	}

	return args
}

// 用于设置互斥参数
func checkMutFlags() error {
	var flagNum int
	var flagStr string
	// 选定 `key:"xx"`
	queryMap, err := utils.StructToMap(args.query, "key")
	if err != nil {
		printer.Error(printer.HandlerLine("Struct To Map err :" + err.Error()))
		return nil
	}
	for k, v := range queryMap {
		if len(v.(string)) != 0 {

			flagNum += 1
		}
		flagStr += k + "/"
	}
	if flagNum == 1 {
		args.Mode = Query_Mode
	}
	queryFileMap, err := utils.StructToMap(args.queryOfFile, "key")
	if err != nil {
		printer.Error(printer.HandlerLine("Struct To Map err :" + err.Error()))
		return nil
	}
	for k, v := range queryFileMap {
		if len(v.(string)) != 0 {
			flagNum += 1
		}
		flagStr += k + "/"
	}
	// 要么单一扫描，要么从文件中批量扫描
	// 单一模式和批量模式互斥
	// 单一模式、批量模式中的各个参数也互斥
	if flagNum > 1 {
		return errors.New("these " + flagStr + " are mutually exclusive")
	}
	// 不输入 query 也应当提醒
	if flagNum == 0 && args.Mode != Stdin_Mode {
		return errors.New("query are empty")
	}
	if args.Mode != Query_Mode {
		args.Mode = File_Mode
	}
	return nil
}

// 检查 email，key
func checkFoFaInfo() {
	if args.FoFaKey == "" || args.FoFaEmail == "" {
		printer.Error("FoFaKey or FoFaEmail is empty")
		os.Exit(1)
	}
}

func getHomeConf() (home string) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "fofa.yaml"
	}
	return filepath.Join(home, ".config", "fofa", "fofa.yaml")
}
