package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bryanlabs/matt/utils/account"
	"github.com/bryanlabs/matt/utils/terraform"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	awsAccountSource = kingpin.Flag("aws-account-source", "Account Source.").Required().Short('a').String()
	tfPath           = kingpin.Arg("tf-path", "Path to terraform modules.").Required().String()
)

// Loop over all profiles and terraform them.
func main() {
	kingpin.Version("0.0.1")
	kingpin.Parse()
	for _, account := range strings.Split(*awsAccountSource, ",") {
		matt(account, *tfPath)
	}
}

// goTerraform will used a named profile to apply a terraform state.
func matt(accountnum string, statepath string) {
	log.Printf("Terraforming %v with %v\n", accountnum, statepath)

	//Update account_id for provider and tfvars.
	providerpath := statepath + "/provider.tf"
	account.UpdateAccountID(providerpath, accountnum)
	tfvarspath := "vars.auto.tfvars"
	account.UpdateAccountID(tfvarspath, accountnum)
	evoaccounts := account.GetEvoAccounts()
	account.UpdateAllAccounts(evoaccounts)
	os.MkdirAll("matt", os.ModePerm)

	// Initialize the new state file
	err := terraform.Init(statepath)

	// Plan the change.
	if err == nil {
		err = terraform.Create(accountnum, statepath)
	}

	if err == nil {

		// Plan the change.
		err = terraform.Apply(accountnum, statepath)
	}
	if err != nil {
		_ = ioutil.WriteFile("matt/"+accountnum+".error.log", []byte(err.Error()), 0)
		log.Printf("### ERROR terraforming %v , See logs for details.\n", accountnum)
	}
	log.Printf("Terraforming %v Complete\n", accountnum)
}
