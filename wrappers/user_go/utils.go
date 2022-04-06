package user_go

import "fmt"

func getFirstAndLastNameWithPrivacy(status NamePrivacyStatus, firstName string, lastName string, userName string) (string, string) {
	if status == NamePrivacyStatusAllHidden || status == NamePrivacyStatusFirstNameHidden ||
		status == NamePrivacyStatusLastNameHidden || (len(firstName) == 0 && len(lastName) == 0) {
		return formatUserName(userName), ""
	}

	return firstName, lastName
}

func formatUserName(userName string) string {
	if len(userName) == 0 {
		return ""
	}

	return fmt.Sprintf("@%v", userName)
}
