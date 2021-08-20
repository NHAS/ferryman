# ferryman

Inspiration taken from: https://github.com/sensepost/reGeorg

Ever wanted to traffic data out of a sensitive environment? Do you like leveraging peoples own infrastructure against them?  
Well now you can! 

Using ferryman, you can open a local port and traffic any data out just as if you were directly connecting to your service. 

Want to catch a regular reverse shell? But cant due to outbound firewall rules? 
Just use ferryman to open a port on the target machines local host, and point your reverse shell to that. Easy beans. 


# Usage

```
ferryman <Service to forward data to> <remote tunnel.aspx path> <local port to open>
ferryman 127.0.0.1:80 https://foobeans.com/tunnel.aspx 5555
```
