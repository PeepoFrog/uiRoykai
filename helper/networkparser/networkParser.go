package networkparser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	interxendpointtypes "github.com/PeepoFrog/km2UI/types/interxEndpoint"
)

var mu sync.Mutex

type Node struct {
	IP      string
	ID      string
	Peers   []Node
	NCPeers int
}

// get nodes that are available by 11000 port
func GetAllNodesV3(ctx context.Context, firstNode string, depth int, ignoreDepth bool) (map[string]Node, error) {
	nodesPool := make(map[string]Node)
	blacklist := make(map[string]string)
	processed := make(map[string]string)
	client := http.DefaultClient
	node, err := GetNetInfoFromInterx(ctx, client, firstNode)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	for _, n := range node.Peers {
		wg.Add(1)
		go loopFunc(ctx, &wg, client, nodesPool, blacklist, processed, n.RemoteIP, 0, depth, ignoreDepth)
	}

	wg.Wait()
	fmt.Println()
	log.Printf("\nTotal saved peers:%v\nOriginal node peer count: %v\nBlacklisted nodes(not reachable): %v\n", len(nodesPool), len(node.Peers), len(blacklist))

	return nodesPool, nil
}

func loopFunc(ctx context.Context, wg *sync.WaitGroup, client *http.Client, pool map[string]Node, blacklist, processed map[string]string, ip string, currentDepth, totalDepth int, ignoreDepth bool) {

	defer wg.Done()
	if !ignoreDepth {
		if currentDepth >= totalDepth {
			// log.Printf("DEPTH LIMIT REACHED")
			return
		}
	}

	// log.Printf("Current depth: %v, IP: %v", currentDepth, ip)

	mu.Lock()
	if _, exist := blacklist[ip]; exist {
		mu.Unlock()
		// log.Printf("BLACKLISTED: %v", ip)
		return
	}
	if _, exist := pool[ip]; exist {
		mu.Unlock()
		// log.Printf("ALREADY EXIST: %v", ip)
		return
	}
	if _, exist := processed[ip]; exist {
		mu.Unlock()
		// log.Printf("ALREADY PROCESSED: %v", ip)
		return
	} else {
		processed[ip] = ip
	}
	mu.Unlock()

	currentDepth++

	var localWaitGroup sync.WaitGroup

	var nodeInfo *interxendpointtypes.NetInfo
	var status *interxendpointtypes.Status
	var errNetInfo error
	var errStatus error
	localWaitGroup.Add(2)
	go func() {
		defer localWaitGroup.Done()
		nodeInfo, errNetInfo = GetNetInfoFromInterx(ctx, client, ip)
	}()
	go func() {
		defer localWaitGroup.Done()
		status, errStatus = GetStatusFromInterx(ctx, client, ip)
	}()

	localWaitGroup.Wait()

	if errNetInfo != nil || errStatus != nil {
		// log.Printf("%v", err.Error())
		mu.Lock()
		log.Printf("adding <%v> to blacklist", ip)
		blacklist[ip] = ip
		cleanValue(processed, ip)
		mu.Unlock()
		// defer localWaitGroup.Done()
		return
	}

	mu.Lock()
	log.Printf("adding <%v> to the pool, nPeers: %v", ip, nodeInfo.NPeers)
	log.Println(status.NodeInfo.ID)

	node := Node{
		IP:      ip,
		NCPeers: nodeInfo.NPeers,
		ID:      status.NodeInfo.ID,
	}
	for _, nn := range nodeInfo.Peers {
		node.Peers = append(node.Peers, Node{IP: nn.RemoteIP, ID: nn.NodeInfo.ID})
	}

	pool[ip] = node
	cleanValue(processed, ip)
	mu.Unlock()

	for _, p := range nodeInfo.Peers {
		wg.Add(1)
		go loopFunc(ctx, wg, client, pool, blacklist, processed, p.RemoteIP, currentDepth, totalDepth, ignoreDepth)
	}

}

func cleanValue(toClean map[string]string, key string) {
	delete(toClean, key)
}

const TimeOutDelay time.Duration = time.Second * 3

func GetNetInfoFromInterx(ctx context.Context, client *http.Client, ip string) (*interxendpointtypes.NetInfo, error) {

	ctxWithTO, c := context.WithTimeout(ctx, TimeOutDelay)
	defer c()
	// log.Printf("Getting net_info from: %v", ip)
	url := fmt.Sprintf("http://%v:11000/api/net_info", ip)
	req, err := http.NewRequestWithContext(ctxWithTO, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeInfo interxendpointtypes.NetInfo
	err = json.Unmarshal(b, &nodeInfo)
	if err != nil {
		return nil, err
	}
	return &nodeInfo, nil
}

func GetStatusFromInterx(ctx context.Context, client *http.Client, ip string) (*interxendpointtypes.Status, error) {
	ctxWithTO, c := context.WithTimeout(ctx, TimeOutDelay)
	defer c()
	// log.Printf("Getting net_info from: %v", ip)
	url := fmt.Sprintf("http://%v:11000/api/status", ip)
	req, err := http.NewRequestWithContext(ctxWithTO, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var nodeStatus interxendpointtypes.Status
	err = json.Unmarshal(b, &nodeStatus)
	if err != nil {
		return nil, err
	}
	return &nodeStatus, nil
}
