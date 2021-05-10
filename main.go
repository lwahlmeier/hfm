package main // import "github.com/lwahlmeier/hfm"

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/PremiereGlobal/stim/pkg/stimlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/mod/semver"
	"gopkg.in/yaml.v2"
)

var log = stimlog.GetLogger()
var version string
var config = viper.New()
var startTime = time.Now().Local()

func init() {

	// Set version for local testing if not set by build system
	lv := "v0.0.0-local"
	if version == "" {
		version = lv
	}
	if !semver.IsValid(version) {
		stimlog.GetLogger().Fatal("Bad Version:{}", version)
	}
}

const (
	PORT       = "port"
	CONFIG     = "config"
	VERSION    = "version"
	SHOWHIDDEN = "showhidden"
)

func main() {
	config.SetEnvPrefix("hfm")
	config.AutomaticEnv()
	config.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))

	var cmd = &cobra.Command{
		Use:   "hfm",
		Short: "launch hfm",
		Long:  "launch hfm",
		Run:   parseInfo,
	}

	cmd.PersistentFlags().Int(PORT, 8844, "port to Listen on (8844 is default)")
	config.BindPFlag(PORT, cmd.PersistentFlags().Lookup(PORT))
	cmd.PersistentFlags().String(CONFIG, "./config.yaml", "config file to use")
	config.BindPFlag(CONFIG, cmd.PersistentFlags().Lookup(CONFIG))
	cmd.PersistentFlags().Bool(VERSION, false, "Get version")
	config.BindPFlag(VERSION, cmd.PersistentFlags().Lookup(VERSION))
	cmd.PersistentFlags().Bool(SHOWHIDDEN, false, "set this flag to show hidden files")
	config.BindPFlag(SHOWHIDDEN, cmd.PersistentFlags().Lookup(SHOWHIDDEN))

	err := cmd.Execute()
	checkError(err)

}

func parseInfo(cmd *cobra.Command, args []string) {
	if config.GetBool(VERSION) {
		fmt.Println("version: " + version)
		return
	}
	port := fmt.Sprintf(":%d", config.GetInt(PORT))
	configFile := config.GetString(CONFIG)
	if _, err := os.Stat(configFile); err != nil {
		log.Fatal("Error trying to check config file:{}", err)
	}
	cba, err := ioutil.ReadFile(configFile)
	checkError(err)
	vds := &VirtualDirs{}
	yaml.Unmarshal(cba, vds)
	log.Info("{} {} {}", port, vds)
	for k, v := range vds.VirtualDirs {
		log.Info("{}:{}", k, v)
	}
	mfs := NewMergedFileSystem(vds.VirtualDirs)
	fs := FileServer(mfs)
	srv := &http.Server{
		Addr:              port,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		IdleTimeout:       10 * time.Second,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Connection", "close")
			fs.ServeHTTP(w, req)
		}),
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			if c2, ok := c.(*net.TCPConn); ok {
				c2.SetWriteBuffer(256 * 1024)
			}
			return ctx
		},
	}
	err = srv.ListenAndServe()
	checkError(err)
}

func checkError(e error) {
	if e != nil {
		log.Fatal("ERROR:{}", e.Error())
	}
}
