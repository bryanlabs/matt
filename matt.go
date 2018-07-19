package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bryanlabs/matt/utils/account"
	"github.com/bryanlabs/matt/utils/auth"
	"github.com/bryanlabs/matt/utils/terraform"
)

// Loop over all profiles and terraform them.
func main() {
	for _, profile := range auth.GetProfiles() {
		err := matt(profile, os.Args[1])
		if err != nil {
			log.Printf("### ERROR terraforming %v , See logs for details.\n", profile.Name)
			continue
		}
	}
}

// goTerraform will used a named profile to apply a terraform state.
func matt(p auth.AWS_Named_Profile, statepath string) error {
	log.Printf("Terraforming %v with %v\n", p.Name, statepath)

	// get the account number from arn.
	arnslice := strings.Split(p.Arn, ":")
	accountnum := arnslice[4]

	//Update account_id for provider and tfvars.
	providerpath := statepath + "/provider.tf"
	account.UpdateAccountID(providerpath, accountnum)
	tfvarspath := statepath + "vars.auto.tfvars"
	account.UpdateAccountID(tfvarspath, accountnum)
	account.UpdateAllAccounts()

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
		_ = ioutil.WriteFile("logs/errors/"+accountnum+".error.log", []byte(err.Error()), 0)
		return err
	}
	log.Printf("Terraforming %v Complete\n", p.Name)
	return err
}
