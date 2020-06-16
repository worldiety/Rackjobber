package rackconfig

import (
	"errors"
)

//GetGitPassFromKeychain retrieves the password matching the given account and domain from the macOS Keychain
func getGitPassFromKeychain(account, domain string) (string, string, error) {
	/*for len(account) == 0 {
		account = rackinput.AwaitTextInput("Username for Git Domain " + domain + "?")
	}

	filledDomain := ""
	if !strings.Contains(domain, "https://") {
		filledDomain = "https://" + domain
	}

	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassInternetPassword)
	query.SetService(filledDomain)
	query.SetAccount(account)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		fmt.Println("Failed to retrieve Authentication from Keychain: ", err)
		return "", "", err
	} else if len(results) == 0 {
		fmt.Printf("No Authentication found for User %v on Domain %v\n", account, domain)
		return "", "", errors.New("no Authentication found")
	}

	password := string(results[0].Data)

	return account, password, nil*/
	return "", "", errors.New("keychain only available on darwin")
}
