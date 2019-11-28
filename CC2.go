package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strings"
)

var nodos = make(map[string]bool)
var RegistrarPuertoHost string
var AgregarPuertoHost string
var nodosChan = make(chan string, 1)

func EnviarSinRespuesta(msg, hostTarget string) {
	host := fmt.Sprintf(":%s", hostTarget)
	conn, err := net.Dial("tcp", host)
	defer conn.Close()
	if err != nil {
		fmt.Printf("Error en enviar notificacion %s", err.Error())
	}

	fmt.Fprintf(conn, "%s\n", msg)
}
func EnviarConRespuesta(hostTarget string) {
	conn, _ := net.Dial("tcp", hostTarget)
	defer conn.Close()
	fmt.Fprintf(conn, "%s\n", AgregarPuertoHost)
	r := bufio.NewReader(conn)
	msgInput, _ := r.ReadString('\n')
	<-nodosChan
	var nodosInput map[string]bool
	json.Unmarshal([]byte(msgInput), &nodosInput)
	for k, v := range nodosInput {
		nodos[k] = v
	}
	fmt.Printf("El mapa actualizado cliente %v", nodos)
	nodosChan <- "Terminado"
}
func AgregarServidor() {
	host := fmt.Sprintf(":%s", AgregarPuertoHost)
	ln, _ := net.Listen("tcp", host)
	defer ln.Close()
	for {
		con, _ := ln.Accept()
		go func(con net.Conn) {
			defer con.Close()
			r := bufio.NewReader(con)
			nodo, _ := r.ReadString('\n')
			<-nodosChan
			nodos[nodo] = true
			fmt.Printf("El mapa actualizado %v", nodos)
			nodosChan <- "Terminado"
		}(con)
	}

}
func ClienteAgregar(nodo string) {
	for k, _ := range nodos {
		k = strings.TrimSpace(k)
		EnviarSinRespuesta(nodo, k)
	}

}
func ClienteRegistrar(hostTarget string) {
	host := fmt.Sprintf(":%s", hostTarget)
	go EnviarConRespuesta(host)
}
func RegistrarServidor() {
	host := fmt.Sprintf(":%s", RegistrarPuertoHost)
	ln, _ := net.Listen("tcp", host)
	defer ln.Close()
	for {
		con, _ := ln.Accept()
		go func(con net.Conn) {
			defer con.Close()
			r := bufio.NewReader(con)
			nodo, _ := r.ReadString('\n')
			ClienteAgregar(nodo)
			<-nodosChan
			var nodoAux = make(map[string]bool)
			for k, v := range nodos {
				nodoAux[k] = v
			}
			nodoAux[hostAggregatorPort] = true
			jsonNodos, _ := json.Marshal(nodoAux)
			fmt.Fprintf(con, "%s\n", string(jsonNodos))
			nodos[nodo] = true
			fmt.Printf("El mapa actualizado %v", nodos)
			nodosChan <- "terminado"

		}(con)
	}

}

func main() {
	nodosChan <- "Inicio"
	ginRegistrador := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto de registro: ")
	RegistrarPuertoHost, _ = ginRegistrador.ReadString('\n')
	RegistrarPuertoHost = strings.TrimSpace(RegistrarPuertoHost)

	ginAgregador := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto agregador: ")
	AgregarPuertoHost, _ = ginAgregador.ReadString('\n')
	AgregarPuertoHost = strings.TrimSpace(AgregarPuertoHost)

	go AgregarServidor()
	go RegistrarServidor()
	ginRemotePort := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto remoto: ")
	remotePort, _ := ginRemotePort.ReadString('\n')
	remotePort = strings.TrimSpace(remotePort)
	if len(remotePort) > 0 {
		ClienteRegistrar(remotePort)
	}
	for {
	}

}
