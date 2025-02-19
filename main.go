package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"userclouds.com/authz"
	"userclouds.com/infra/jsonclient"
	"userclouds.com/infra/pagination"
)

func runAuthZ() {
	ctx := context.Background()
	clientID := os.Getenv("USERCLOUDS_CLIENT_ID")
	clientSecret := os.Getenv("USERCLOUDS_CLIENT_SECRET")
	region := os.Getenv("USERCLOUDS_REGION")
	if region == "" {
		region = "eu-west-1"
	}
	tenantHost := "usercloudstests-connectivity-tests.tenant.userclouds.com"
	tenantURL := fmt.Sprintf("https://%s", tenantHost)
	endpointURL := fmt.Sprintf("https://aws-%s-eks.userclouds.com", region)
	log.Printf("Using URL: %s", endpointURL)
	if clientID == "" || clientSecret == "" {
		log.Fatal("missing one or more required environment variables: USERCLOUDS_CLIENT_ID, USERCLOUDS_CLIENT_SECRET")
	}

	ts, err := jsonclient.ClientCredentialsForURL(tenantURL, clientID, clientSecret, nil)
	if err != nil {
		log.Fatalf("Error getting token. check values of provided client id, client secret and tenant url : %s", err)
	}
	client, err := authz.NewClient(endpointURL, authz.BypassCache(), authz.JSONClient(ts), authz.JSONClient(jsonclient.HeaderHost(tenantHost)))

	if err != nil {
		log.Fatalf("Error creating authz client: %s", err)
	}

	for {
		authObjects(ctx, client)
		authZEdges(ctx, client)
	}
}

func authZEdges(ctx context.Context, client *authz.Client) {
	edges := enumerateEdges(ctx, client)
	for i, edge := range edges {

		_, err := client.GetObject(ctx, edge.SourceObjectID)
		if err != nil {
			log.Fatalf("Error getting source object for edge %d: %s", i, err)
		}
		_, err = client.GetObject(ctx, edge.TargetObjectID)
		if err != nil {
			log.Fatalf("Error getting target object for edge %d: %s", i, err)
		}
		et, err := client.GetEdgeType(ctx, edge.EdgeTypeID)
		if err != nil {
			log.Fatalf("Error getting edge type for edge %d: %s", i, err)
		}
		log.Printf("Got objects for edge %d: %v [type: %v]", i, edge.ID, et.TypeName)
	}
}

func authObjects(ctx context.Context, client *authz.Client) {
	objects := enumerateObjects(ctx, client)
	for i, object := range objects {
		objType, err := client.GetObjectType(ctx, object.TypeID)
		if err != nil {
			log.Fatalf("Error getting object type for object %d: %s", i, err)
		}
		log.Printf("Got object %d: %v [type: %v]", i, object.ID, objType.TypeName)
	}
}
func enumerateEdges(ctx context.Context, client *authz.Client) []authz.Edge {
	cursor := pagination.CursorBegin
	edges := make([]authz.Edge, 0)
	for {
		es, err := client.ListEdges(ctx, authz.Pagination(pagination.StartingAfter(cursor)))
		if err != nil {
			log.Fatalf("Error listing edges: %s", err)
		}
		edges = append(edges, es.Data...)
		if !es.HasNext {
			break
		}
		log.Printf("Got %d edges, next: %s", len(edges), es.Next)
		cursor = es.Next
	}
	log.Printf("Got %d edges", len(edges))
	return edges
}

func enumerateObjects(ctx context.Context, client *authz.Client) []authz.Object {
	cursor := pagination.CursorBegin
	objects := make([]authz.Object, 0)
	for {
		objs, err := client.ListObjects(ctx, authz.Pagination(pagination.StartingAfter(cursor)))
		if err != nil {
			log.Fatalf("Error listing objects: %s", err)
		}
		objects = append(objects, objs.Data...)
		if !objs.HasNext {
			break
		}
		log.Printf("Got %d objects, next: %s", len(objects), objs.Next)
		cursor = objs.Next
	}
	log.Printf("Got %d objects", len(objects))
	return objects
}

func main() {
	runAuthZ()

}
