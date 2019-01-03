package dynamicdiscovery

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	discclient "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/discovery/client"
	fabdiscovery "github.com/hyperledger/fabric-sdk-go/internal/github.com/hyperledger/fabric/protos/discovery"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	fabcontext "github.com/hyperledger/fabric-sdk-go/pkg/common/providers/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/context"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/discovery"
	"github.com/hyperledger/fabric-sdk-go/third_party/github.com/hyperledger/fabric/protos/msp"
	"github.com/pkg/errors"
	"strings"
	"time"
)

const (
	peer0Org1    = "peer0.org1.example.com"
	orgChannelID = "mychannel"
)

func newCCCall(ccID string, collections ...string) *fabdiscovery.ChaincodeCall {
	return &fabdiscovery.ChaincodeCall{
		Name:            ccID,
		CollectionNames: collections,
	}
}

func GetEndorsers(ctx fabcontext.Client, client *discovery.Client, peerConfig fab.PeerConfig) {

	interest := newInterest(newCCCall("mycc"))
	_, err := retry.NewInvoker(retry.New(retry.TestRetryOpts)).Invoke(
		func() (interface{}, error) {
			chanResp, err := sendEndorserQuery(ctx, client, interest, peerConfig)
			if err != nil && strings.Contains(err.Error(), "failed constructing descriptor for chaincodes") {
				return nil, errors.WithMessage(err, "failed constructing descriptor for chaincodes")
			} else if err != nil {
				return nil, errors.WithMessage(err, "Got error from discovery")
			}

			return chanResp, nil
		},
	)
	if err != nil {
		fmt.Errorf("==== sendEndorserQuery faild! err:%s\n", err)
	}

}

func sendEndorserQuery(ctx fabcontext.Client, client *discovery.Client, interest *fabdiscovery.ChaincodeInterest, peerConfig fab.PeerConfig) (discclient.ChannelResponse, error) {
	req, err := discclient.NewRequest().OfChannel(orgChannelID).AddEndorsersQuery(interest)

	reqCtx, cancel := context.NewRequest(ctx, context.WithTimeout(10*time.Second))
	defer cancel()

	responses, err := client.Send(reqCtx, req, peerConfig)

	chanResp := responses[0].ForChannel(orgChannelID)

	endorsers, err := chanResp.Endorsers(interest.Chaincodes, discclient.NewFilter(discclient.NoPriorities, discclient.NoExclusion))
	if err != nil {
		return nil, err
	}
	fmt.Printf("*******************************\n")
	//asURLs(endorsers)
	PrintPeerInfo(endorsers)

	fmt.Printf("*******************************\n")

	return chanResp, nil
}

func PrintPeerInfo(peers []*discclient.Peer) {
	for _, endorser := range peers {
		aliveMsg := endorser.AliveMessage.GetAliveMsg()
		fmt.Printf("---- aliveMsg.Membership.Endpoint:[%s]\n", aliveMsg.Membership.Endpoint)

		sID := &msp.SerializedIdentity{}
		if err := proto.Unmarshal(endorser.Identity, sID); err != nil {
			panic(fmt.Errorf("failed unmarshaling peer's identity, err:%s \n", err))
		}

		fmt.Printf(" ---- endorser.Mspid:[%#+v] \n", sID.Mspid)
		fmt.Printf(" ---- endorser.cert:[%#+v] \n", string(sID.IdBytes))

		stateInfo := endorser.StateInfoMessage.GetStateInfo()

		fmt.Printf("--- Ledger Height: %d \n", stateInfo.Properties.LedgerHeight)
		fmt.Printf("--- Chaincodes: \n")
		for _, cc := range stateInfo.Properties.Chaincodes {
			fmt.Printf("------ %s:%s \n", cc.Name, cc.Version)
		}

	}
}

func asURLs(endorsers discclient.Endorsers) []string {
	var urls []string
	fmt.Printf("len(endorsers):[%d]\n", len(endorsers))
	for _, endorser := range endorsers {
		aliveMsg := endorser.AliveMessage.GetAliveMsg()
		urls = append(urls, aliveMsg.Membership.Endpoint)
		fmt.Printf("---- aliveMsg.Membership.Endpoint:[%s]\n", aliveMsg.Membership.Endpoint)

		sID := &msp.SerializedIdentity{}
		if err := proto.Unmarshal(endorser.Identity, sID); err != nil {
			fmt.Errorf("failed unmarshaling peer's identity, err:%s \n", err)
			return nil
		}

		fmt.Printf(" endorser.Identity:[%#+v] \n", sID.Mspid)

		fmt.Printf("---- endorser.AliveMessage.Content:[%#+v]\n", endorser.AliveMessage.Content)

	}
	return urls
}

func checkEndorsers(endorsers []string, expectedGroups [][]string) {
	for _, group := range expectedGroups {
		if containsAll(endorsers, group) {
			fmt.Printf("Found matching endorser group: %#v", group)
			return
		}
	}
}

func containsAll(endorsers []string, expectedEndorserGroup []string) bool {
	if len(endorsers) != len(expectedEndorserGroup) {
		return false
	}

	for _, endorser := range endorsers {
		fmt.Printf("Checking endpoint: %s ...", endorser)
		if !contains(expectedEndorserGroup, endorser) {
			return false
		}
	}
	return true
}

func contains(group []string, endorser string) bool {
	for _, e := range group {
		if e == endorser {
			return true
		}
	}
	return false
}
func newInterest(ccCalls ...*fabdiscovery.ChaincodeCall) *fabdiscovery.ChaincodeInterest {
	return &fabdiscovery.ChaincodeInterest{Chaincodes: ccCalls}
}
