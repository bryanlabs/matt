package main

import (
	"log"
	"os"
	"strings"
	"sync"

	"github.com/bryanlabs/matt/utils/account"
	"github.com/bryanlabs/matt/utils/terraform"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	awsAccountSource = kingpin.Flag("aws-account-source", "Account Source.").Required().Short('a').String()
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
	for _, account := range slice {
		matt(account, *tfCmd, *tfPath)
	}
	wg.Wait()
}

// goTerraform will used a named profile to apply a terraform state.
func matt(accountnum string, tfcmd string, statepath string) {
	log.Printf("Terraforming %v with %v\n", accountnum, statepath)
	modulename := account.GetModuleName(statepath)
	//Update account_id for provider and tfvars.
	providerpath := statepath + "/provider.tf"
	account.UpdateAccountID(providerpath, accountnum, modulename)
	tfvarspath := "vars.auto.tfvars"
	account.UpdateAccountID(tfvarspath, accountnum, modulename)
	evoaccounts := account.GetEvoAccounts()
	account.UpdateAllAccounts(evoaccounts)
	os.MkdirAll("matt", os.ModePerm)
	if len(tfcmd) > 1 {
		switch cmd := tfcmd; cmd {
		case "create":
			terraform.Init(statepath)
			terraform.Create(accountnum, statepath)
		case "apply":
			terraform.Init(statepath)
			terraform.Apply(accountnum, statepath)
		case "destroy":
			terraform.Init(statepath)
			terraform.Destroy(accountnum, statepath)
		default:
			log.Printf("Command: %v not supported\n", tfcmd)
		}
	} else {
		terraform.Init(statepath)
		terraform.Create(accountnum, statepath)
		terraform.Apply(accountnum, statepath)
	}

	log.Printf("Terraforming %v Complete\n", accountnum)
	wg.Done()
}
