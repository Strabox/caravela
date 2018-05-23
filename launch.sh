interface="$(route | grep '^default' | grep -o '[^ ]*$')"
echo $interface
ip=$(ifconfig | grep -w "$interface")
echo $ip