# digest-proxy

## About
This is a simple proxy server to proxy the request to the target webserver with digest auth


## How to run

Download the binary, set environment variables and start the proxy server
```
export TARGET_URL='https://api.example.com:8080'   // the target web service address you want to request
export USERNAME='someuser'                         // user name
export PASSWORD='user-password'                    // password
export PORT='3000'                                 // proxy server listen port
./digest-proxy
```

I also provide a docker image to simplify the deployment, you may launch the proxy server using the following command
```
docker run -d -e TARGET_URL='https://api.example.com' -e USERNAME="someuser" -e PASSWORD='password' -e PORT='3000' -p 3000:3000  wizardonmoon/digest-proxy
```
