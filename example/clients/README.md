

```bash

./aporeto_shell.bash
./client_kinit_client_write.bash

Get token->
{
  "service": {
    "keytabs": "superman:birdman",
    "scopes": "default",
    "secrets": "secret1:secret2"
  },
  "aud": "initial",
  "exp": 1604953861,
  "iat": 1604950261,
  "iss": "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
  "sub": "5fa998edc3a26d00019b7a6a"
}

Get nonce with token from above->
{
  "exp": 1604950321,
  "value": "85T2KsuYMDiJn9gC4uhhi6Ohy67wnoLjdSGBwr81kjbxHoYcI24F4lBTu1116Hbd"
}

Get token with audience (aud field) set to nonce 85T2KsuYMDiJn9gC4uhhi6Ohy67wnoLjdSGBwr81kjbxHoYcI24F4lBTu1116Hbd->
{
  "service": {
    "keytabs": "superman:birdman",
    "scopes": "default",
    "secrets": "secret1:secret2"
  },
  "aud": "85T2KsuYMDiJn9gC4uhhi6Ohy67wnoLjdSGBwr81kjbxHoYcI24F4lBTu1116Hbd",
  "exp": 1604953861,
  "iat": 1604950261,
  "iss": "https://api.console.aporeto.com/v/1/namespaces/5ddc396b9facec0001d3c886/oauthinfo",
  "sub": "5fa998edc3a26d00019b7a6a"
}
Get keytab with token from above and NAME superman->
{
  "principal": "HTTP/superman@EXAMPLE.COM",
  "base64file": "BQIAAABMAAIAC0VYQU1QTEUuQ09NAARIVFRQAAhzdXBlcm1hbgAAAAEAAAAAAQASACAwtH0XXSb/V9+PDVbZvck57U73YWcbla6/HHkNDI47gg==",
  "exp": 1604950320
}
Authenticate with Active Directory / Kerberos Server: Successful
Mount Windows CIFS share with user superman: Successful
Write random message This is my random message Qktti+Nm3ERh98z9drRJGbtD to /data/random.txt: Successful
```
