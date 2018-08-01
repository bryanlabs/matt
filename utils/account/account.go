package account

import (
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/guregu/dynamo"
)

type AWS_dbaccountinfo struct {
	AccountID     string
	Project       string
	IsDevelopment bool
	Deleted_at    string
}

func GetModuleName(path string) (modulename string) {
	slice := strings.Split(path, "\\")
	li := slice[len(slice)-2]
	return li
}

// updateAccountID will modify provider.tf and dynamic.tfvars to allow multi account terraforming.
func UpdateAccountID(file string, account string, module string) {
	data, err := ioutil.ReadFile(file)
	check(err)
	lines := strings.Split(string(data), "\n")
	for i, line := range lines {
		if strings.Contains(line, "tfstate") {
			lines[i] = "    key    = \"" + account + "-" + module + ".tfstate\""
		} else if strings.HasPrefix(line, "account_id =") {
			lines[i] = "account_id = \"" + account + "\""
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(file, []byte(output), 0)
}

func GetEvoAccounts() (evoaccounts []string) {
	// Create DynamoDB client

	db := dynamo.New(session.New(), &aws.Config{Region: aws.String("us-east-1")})
	table := db.Table("AccountSecurityGroups")
	var results []AWS_dbaccountinfo

	err := table.Scan().All(&results)

	if err != nil {
		log.Fatal("Could not get session: ", err)
	}

	for _, r := range results {
		if r.Deleted_at == "" && r.IsDevelopment != true {
			evoaccounts = append(evoaccounts, r.AccountID)
		}
	}
	return
}

type AccountInfo struct {
	Arns     string
	Accounts string
}

func GetAccountInfo(allAccounts []string) (info AccountInfo) {

	// Update the tfvars file.
	var accountARNs []string
	var accountNumbers []string
	for _, account := range allAccounts {
		//Need the tripple quotes for powershell.
		ARN := "\"\"\"arn:aws:sts::" + account + ":assumed-role/lambda-ssmhealth/ssmhealth\"\"\""
		accountNumbers = append(accountNumbers, "\"\"\""+account+"\"\"\"")
		accountARNs = append(accountARNs, ARN)
	}
	info.Arns = "[" + strings.Join(accountARNs, ", ") + "]"
	info.Accounts = "[" + strings.Join(accountNumbers, ", ") + "]"
	return
}

func check(e error) {
	if e != nil {
		log.Printf("Error: %v\n", e)
		return
	}
}
