package gmailtest

func CredentialsJSON() []byte {
	return []byte(`{
  "installed": {
    "client_id": "888888888888-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx.apps.googleusercontent.com",
    "project_id": "test",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
    "redirect_uris": [
      "http://localhost"
    ]
  }
}`)
}
