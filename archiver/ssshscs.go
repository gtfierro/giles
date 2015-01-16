package archiver

import (
	"bufio"
	"bytes"
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io"
	"io/ioutil"
	"net"
	"os"
	"strings"
)

type SSHConfigServer struct {
	store              *Store
	port               string
	authorizedKeysFile string
	config             *ssh.ServerConfig
}

func NewSSHConfigServer(store *Store, port, privatekey, authorizedKeysFile, confuser, confpass string, passenabled, keyenabled bool) *SSHConfigServer {

	keys := loadkeys(authorizedKeysFile)
	config := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
			if passenabled {
				if c.User() == confuser && string(pass) == confpass {
					return nil, nil
				}
			}
			return nil, fmt.Errorf("password rejected for %q", c.User())
		},

		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			if keyenabled {
				for _, authkey := range keys {
					if bytes.Compare(authkey.Marshal(), key.Marshal()) == 0 {
						return nil, nil
					}
				}
			}
			return nil, fmt.Errorf("Publickey authorization failed")
		},
	}

	privateBytes, err := ioutil.ReadFile(privatekey)
	if err != nil {
		log.Fatal("Failed to load private key %v", privatekey)
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)
	sshscs := &SSHConfigServer{store: store,
		port:               port,
		authorizedKeysFile: authorizedKeysFile,
		config:             config}
	return sshscs
}

func (scs *SSHConfigServer) Listen() {
	listener, err := net.Listen("tcp", "0.0.0.0:"+scs.port)
	if err != nil {
		log.Fatalf("Failed to listen on port %v (%s)", scs.port, err)
	}
	log.Info("Listening on %v...", scs.port)
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Error("Failed to accept incoming connection (%s)", err)
			continue
		}
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, scs.config)
		if err != nil {
			log.Error("Failed to handshake (%s)", err)
			continue
		}

		log.Notice("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		go scs.handleRequests(reqs)
		go scs.handleChannels(chans)
	}
}

func loadkeys(filename string) []ssh.PublicKey {
	ret := []ssh.PublicKey{}
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		log.Error("Failed to open authorized_keys file (%s)", err)
		return ret
	}

	reader := bufio.NewReader(file)
	for {
		line, err := reader.ReadBytes('\n') // read up until newline
		if err == io.EOF {                  // done
			break
		} else if err != nil {
			log.Error("Error reading authorized_keys file (%s)", err)
			return ret
		}
		pub, _, _, _, err := ssh.ParseAuthorizedKey(line)
		if err != nil {
			log.Error("Error parsing key from line %s (%s)", line, err)
			continue
		}
		ret = append(ret, pub)
	}
	return ret
}

func (scs *SSHConfigServer) handleRequests(requests <-chan *ssh.Request) {
	for req := range requests {
		log.Info("OOB OOB %+v", req)
	}
}

func (scs *SSHConfigServer) handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		log.Info("Channel type %s", newChannel.ChannelType())
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Error("Could not accept channel (%s)", err)
			continue
		}

		scs.handleChannel(channel, requests)
	}
}

func (scs *SSHConfigServer) handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	go func(in <-chan *ssh.Request) {
		for req := range requests {
			switch req.Type {
			case "shell":
				if len(req.Payload) == 0 {
					req.Reply(true, nil)
				} else { // got an external command
					log.Info("shell payload %s", req.Payload)
				}
			case "pty-req":
				log.Info("ptyreq")
				req.Reply(true, nil)
			case "window-change":
				log.Info("window chnge")
			default:
				log.Info("type: %s, payload %s", req.Type, req.Payload)
			}
		}
	}(requests)

	term := terminal.NewTerminal(channel, "giles> ")

	go func() {
		scs.writeLines(term, greeting)
		defer channel.Close()
		for {
			line, err := term.ReadLine()
			if err != nil {
				break
			}
			fmt.Println(line)
			scs.handleInput(term, &channel, line)
		}
	}()
}

func (scs *SSHConfigServer) handleInput(term *terminal.Terminal, channel *ssh.Channel, line string) {
	switch {
	case line == "quit":
		term.Write([]byte("Quitting!\r\n"))
		(*channel).Close()
	case line == "help":
		scs.writeLines(term, help)
	case strings.HasPrefix(line, "newkey"):
		key := scs.newkey(line)
		scs.writeLines(term, key)
	case strings.HasPrefix(line, "getkey"):
		key := scs.getkey(line)
		scs.writeLines(term, key)
	case strings.HasPrefix(line, "listkeys"):
		keys := scs.listkeys(line)
		scs.writeLines(term, keys)
	case strings.HasPrefix(line, "delkey"):
		success := scs.delkey(line)
		scs.writeLines(term, success)
	case strings.HasPrefix(line, "owner"):
		owner := scs.owner(line)
		scs.writeLines(term, owner)
	default:
		term.Write([]byte("default\r\n"))
	}
}

func (scs *SSHConfigServer) writeLines(term *terminal.Terminal, lines string) {
	for _, line := range strings.Split(lines, "\n") {
		term.Write([]byte(line + "\r\n"))
	}
}

func (scs *SSHConfigServer) newkey(line string) string {
	var response string
	return response
}

func (scs *SSHConfigServer) getkey(line string) string {
	var response string
	return response
}

func (scs *SSHConfigServer) listkeys(line string) string {
	var response string
	return response
}

func (scs *SSHConfigServer) delkey(line string) string {
	var response string
	return response
}

func (scs *SSHConfigServer) owner(line string) string {
	var response string
	return response
}

var greeting = `
Welcome to SSSHSCS, the sMAP SSH Server Configuration Shell!
     ______   ___   ___    _____  ________  ______
    |\   ___\|\  \ |\  \  / __  \|\   __  \|\___   \
    \ \  \__|\ \  \\_\  \|\/_|\  \ \  \|\  \|___|\  \
     \ \  \   \ \______  \|/ \ \  \ \  \\\  \   \ \  \
A     \ \  \___\|_____|\  \   \ \  \ \  \\\  \  _\_\  \  production
       \ \______\     \ \__\   \ \__\ \_______\|\______\
        \|______|      \|__|    \|__|\|_______|\|______|
`

var help = `
Welcome to SSSHSCS, the sMAP SSH Server Configuration Shell!

We support the following commands:

[[General]]
quit -- exits the session
help -- prints this help

[[Key Management]]
newkey <name> <email> <public?> -- creates a new API key and prints it
getkey <name> <email> -- retrieve the API key for the given name and email
listkeys <email> -- list all API keys and names for the given email
delkey <name> <email> -- deletes the key associated with the given name and
						 email
delkey <key> -- deletes the given key
owner <key> -- retrieves owner (name, email) for given key
`
