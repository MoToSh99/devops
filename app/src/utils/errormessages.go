package utils

const (
	//UsernameTaken Error message for when username is already taken.
	UsernameTaken = "The username is already taken"
	//PasswordDoesNotMatchMessage Error message for when the user failed to provide the same password twice.
	PasswordDoesNotMatchMessage = "The two passwords do not match"
	//YouHaveToEnterAPassword Error message for when no password was provided.
	YouHaveToEnterAPassword = "You have to enter a password" /* #nosec G101 */
	//EnterAValidEmail Error message for when the provided email address is not valid.
	EnterAValidEmail = "You have to enter a valid email address"
	//EnterAUsername Error message for when no username was provided
	EnterAUsername = "You have to enter a username"
	//PasswordMustBeAtleast8Chars Error message when the input password consist of less than 8 characters
	PasswordMustBeAtleast8Chars = "Your password must consist of atleast 8 characters"
	//PasswordMustContainAtleastOneUppercase Error message when the input password does not contain atleast one uppercase letter.
	PasswordMustContainAtleastOneUppercase = "Your password must include atleast 1 uppercase letter"
)
