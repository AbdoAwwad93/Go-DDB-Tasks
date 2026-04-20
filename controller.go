package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
)

type Request struct {
	Command   string `json:"command"`
	Wallpaper string `json:"wallpaper,omitempty"`
}

type Response struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type Agent struct {
	Name    string
	Conn    net.Conn
	Encoder *json.Encoder
	Decoder *json.Decoder
	mu      sync.Mutex
}

func main() {
	port := flag.Int("port", 9000, "controller listen port")
	flag.Parse()

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen on port %d: %v", *port, err)
	}
	defer listener.Close()

	log.Printf("controller listening on :%d", *port)
	log.Println(`type "lock", "shutdown", "wallpaper C:\path\image.jpg", "list", or "exit"`)

	var (
		agentsMu sync.RWMutex
		agents   = make(map[string]*Agent)
	)

	go acceptAgents(listener, &agentsMu, agents)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			log.Println("input closed, shutting down controller")
			return
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch strings.ToLower(line) {
		case "exit", "quit":
			log.Println("controller stopped")
			return
		case "list":
			printAgents(&agentsMu, agents)
			continue
		}

		req, err := parseCommand(line)
		if err != nil {
			log.Printf("invalid command: %v", err)
			continue
		}

		broadcast(req, &agentsMu, agents)
	}
}

func acceptAgents(listener net.Listener, agentsMu *sync.RWMutex, agents map[string]*Agent) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("accept error: %v", err)
			continue
		}

		go registerAgent(conn, agentsMu, agents)
	}
}

func registerAgent(conn net.Conn, agentsMu *sync.RWMutex, agents map[string]*Agent) {
	decoder := json.NewDecoder(conn)

	var hello map[string]string
	if err := decoder.Decode(&hello); err != nil {
		log.Printf("agent handshake failed from %s: %v", conn.RemoteAddr(), err)
		_ = conn.Close()
		return
	}

	name := strings.TrimSpace(hello["name"])
	if name == "" {
		name = conn.RemoteAddr().String()
	}

	agentID := conn.RemoteAddr().String()
	agent := &Agent{
		Name:    name,
		Conn:    conn,
		Encoder: json.NewEncoder(conn),
		Decoder: decoder,
	}

	agentsMu.Lock()
	agents[agentID] = agent
	agentsMu.Unlock()

	log.Printf("agent connected: %s (%s)", name, agentID)
}

func printAgents(agentsMu *sync.RWMutex, agents map[string]*Agent) {
	agentsMu.RLock()
	defer agentsMu.RUnlock()

	if len(agents) == 0 {
		fmt.Println("no agents connected")
		return
	}

	fmt.Println("connected agents:")
	for id, agent := range agents {
		fmt.Printf("- %s (%s)\n", agent.Name, id)
	}
}

func parseCommand(line string) (Request, error) {
	parts := strings.SplitN(strings.TrimSpace(line), " ", 2)
	command := strings.ToLower(strings.TrimSpace(parts[0]))

	req := Request{Command: command}
	switch command {
	case "lock", "shutdown":
		return req, nil
	case "wallpaper":
		if len(parts) < 2 || strings.TrimSpace(parts[1]) == "" {
			return Request{}, fmt.Errorf("wallpaper command requires a path")
		}

		req.Wallpaper = strings.TrimSpace(parts[1])
		return req, nil
	default:
		return Request{}, fmt.Errorf("supported commands are: lock, shutdown, wallpaper <path>, list, exit")
	}
}

func broadcast(req Request, agentsMu *sync.RWMutex, agents map[string]*Agent) {
	agentsMu.RLock()
	snapshot := make(map[string]*Agent, len(agents))
	for id, agent := range agents {
		snapshot[id] = agent
	}
	agentsMu.RUnlock()

	if len(snapshot) == 0 {
		fmt.Println("no agents connected")
		return
	}

	var wg sync.WaitGroup
	results := make(chan string, len(snapshot))

	for id, agent := range snapshot {
		wg.Add(1)
		go func(agentID string, a *Agent) {
			defer wg.Done()
			results <- sendCommand(a, agentID, req, agentsMu, agents)
		}(id, agent)
	}

	wg.Wait()
	close(results)

	fmt.Println("broadcast results:")
	for result := range results {
		fmt.Println(result)
	}
}

func sendCommand(agent *Agent, agentID string, req Request, agentsMu *sync.RWMutex, agents map[string]*Agent) string {
	agent.mu.Lock()
	defer agent.mu.Unlock()

	if err := agent.Encoder.Encode(req); err != nil {
		removeAgent(agentID, agent, agentsMu, agents)
		return fmt.Sprintf("[%s] send failed: %v", agent.Name, err)
	}

	var resp Response
	if err := agent.Decoder.Decode(&resp); err != nil {
		removeAgent(agentID, agent, agentsMu, agents)
		return fmt.Sprintf("[%s] response failed: %v", agent.Name, err)
	}

	if resp.OK {
		return fmt.Sprintf("[%s] success: %s", agent.Name, resp.Message)
	}

	return fmt.Sprintf("[%s] failed: %s", agent.Name, resp.Message)
}

func removeAgent(agentID string, agent *Agent, agentsMu *sync.RWMutex, agents map[string]*Agent) {
	_ = agent.Conn.Close()

	agentsMu.Lock()
	delete(agents, agentID)
	agentsMu.Unlock()
}
