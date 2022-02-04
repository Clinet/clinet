(go1.18beta1 build -ldflags="-s -w" -o clinet.exe || (Write-Error 'Failed to build Clinet.' -ErrorAction Stop)) && .\clinet.exe -debug true -bot true -gcptoken google-assistant-token.json
