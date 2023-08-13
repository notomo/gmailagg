package gcstest

func CredentialsJSON() []byte {
	return []byte(`{
  "client_id": "xxxxxxxxxxxx-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
  "client_secret": "x-xxxxxxxxxxxxxxxxxxxxxx",
  "quota_project_id": "test",
  "refresh_token": "1//XXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
  "type": "authorized_user"
}`)
}
