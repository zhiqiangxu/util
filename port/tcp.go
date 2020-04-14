package port

import "net"

// ChooseTCP returns a useable tcp port
func ChooseTCP() (p uint16, err error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return
	}

	p = uint16(l.Addr().(*net.TCPAddr).Port)

	err = l.Close()
	return
}

// ListenTCP returns a new tcp listener
func ListenTCP() (l net.Listener, err error) {
	l, err = net.Listen("tcp", ":0")
	return
}
