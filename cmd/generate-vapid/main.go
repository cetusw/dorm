package main

import (
	"fmt"
	"log"

	webpush "github.com/SherClockHolmes/webpush-go"
)

func main() {
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	if err != nil {
		log.Fatalf("generate VAPID keys: %v", err)
	}

	fmt.Printf("WEB_PUSH_PUBLIC_KEY=%s\n", publicKey)
	fmt.Printf("WEB_PUSH_PRIVATE_KEY=%s\n", privateKey)
}
