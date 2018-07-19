package matt

import (
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/bryanlabs/matt/auth"
)

// Loop over all profiles and terraform them.
func main() {
	for _, profile := range auth.getProfiles() {
		err := MATT(profile, os.Args[1])
		if err != nil {
			log.Printf("### ERROR terraforming %v , See logs for details.\n", profile.name)
			continue
		}
	}
}

// goTerraform will used a named profile to apply a terraform state.
func goMATT(p profile, statepath string) error {
	log.Printf("Terraforming %v with %v\n", p.name, statepath)

	// get the account number from arn.
	arnslice := strings.Split(p.arn, ":")
	account := arnslice[4]

	//Update account_id for provider and tfvars.
	providerpath := statepath + "/provider.tf"
	updateAccountID(providerpath, account)
	tfvarspath := "vars.auto.tfvars"
	updateAccountID(tfvarspath, account)
	updateAllAccounts()

	// Initialize the new state file
	err := tfInit(statepath)

	// Plan the change.
	if err == nil {
		err = tfCreatePlan(account, statepath)
	}

	if err == nil {
		// Plan the change.
		err = tfApplyPlan(account, statepath)
	}
	if err != nil {
		_ = ioutil.WriteFile("logs/errors/"+account+".error.log", []byte(err.Error()), 0)
		return err
	}
	log.Printf("Terraforming %v Complete\n", p.name)
	return err
}
