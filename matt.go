package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bryanlabs/matt/utils/account"
	"github.com/bryanlabs/matt/utils/terraform"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/ini.v1"
)

var (
	awsAccountSource = kingpin.Flag("aws-account-source", "Account Source.").Required().Short('a').String()
	mattConf         = kingpin.Flag("conf", "Conf File.").Required().Short('f').String()
	tfCmd            = kingpin.Flag("tf-cmd", "Terraform Command.").Short('c').String()
	tfPath           = kingpin.Arg("tf-path", "Path to terraform modules.").Required().String()
	wg               sync.WaitGroup
)

// Loop over all profiles and terraform them.
func main() {

	kingpin.Version("0.0.1")
	kingpin.Parse()
	slice := strings.Split(*awsAccountSource, ",")
	wg.Add(len(slice))
	os.MkdirAll("matt", os.ModePerm)
	for _, i := range slice {
		matt(i, *tfCmd, *tfPath, *mattConf)
	}
	wg.Wait()

}

// goTerraform will used a named profile to apply a terraform state.
func matt(accountnum string, tfcmd string, statepath string, conf string) {
	cfg, err := ini.Load(conf)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}
	var slice []string
	keys := cfg.Section("matt").KeyStrings()
	for _, key := range keys {
		val := cfg.Section("matt").Key(key).String()
		slice = append(slice, key+"="+val+"")
	}

	log.Printf("Terraforming %v with %v\n", accountnum, statepath)

	//Update account_id for provider and tfvars.
	modulename := account.GetModuleName(statepath)
	evoaccounts := account.GetEvoAccounts()
	info := account.GetAccountInfo(evoaccounts)
	arnstr := "all_accountarns=" + info.Arns
	accountsstr := "all_accountnumbers=" + info.Accounts
	accountstr := "account_id=" + accountnum

	slice = append(slice, arnstr, accountstr, accountsstr)
	options := "-var '" + strings.Join(slice, "' -var '") + "'"
	domain := strings.ToLower(os.Getenv("userdomain"))
	bucket := domain + "-terraform-remote-state-storage-s3"

	if len(tfcmd) > 1 {
		switch cmd := tfcmd; cmd {
		case "create":
			terraform.Init(statepath, options, accountnum, modulename, bucket)
			terraform.Create(accountnum, statepath, options, modulename)
		case "apply":
			terraform.Init(statepath, options, accountnum, modulename, bucket)
			terraform.Apply(accountnum, statepath, options, modulename)
		case "destroy":
			terraform.Init(statepath, options, accountnum, modulename, bucket)
			terraform.Destroy(accountnum, statepath, options, modulename)
		default:
			log.Printf("Command: %v not supported\n", tfcmd)
		}
	} else {
		terraform.Init(statepath, options, accountnum, modulename, bucket)
		terraform.Create(accountnum, statepath, options, modulename)
		terraform.Apply(accountnum, statepath, options, modulename)
	}

	log.Printf("Terraforming %v Complete\n", accountnum)
	wg.Done()
}
