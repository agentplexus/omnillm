package main

import (
	"fmt"
	"log"

	"github.com/agentplexus/omnillm"
)

func main() {
	fmt.Println("=== OmniLLM Provider Architecture Demo ===")
	fmt.Println()

	// Show the current architecture
	fmt.Println("Current Architecture:")
	fmt.Println("📁 omnillm/ (main package)")
	fmt.Println("  ├── client.go        - ChatClient wrapper")
	fmt.Println("  ├── provider.go      - Provider interface")
	fmt.Println("  ├── providers.go     - Provider adapters")
	fmt.Println("  ├── types.go         - Unified types")
	fmt.Println("  ├── errors.go        - Error handling")
	fmt.Println("  └── providers/")
	fmt.Println("      ├── openai/      - OpenAI implementation")
	fmt.Println("      ├── anthropic/   - Claude implementation")
	fmt.Println("      ├── bedrock/     - AWS Bedrock implementation")
	fmt.Println("      └── ollama/      - Ollama local models")
	fmt.Println()

	// Demonstrate creating clients for different providers
	fmt.Println("Creating clients for different providers...")

	// OpenAI client (won't work without real API key)
	openaiClient, err := omnillm.NewClient(omnillm.ClientConfig{
		Providers: []omnillm.ProviderConfig{
			{Provider: omnillm.ProviderNameOpenAI, APIKey: "demo-openai-key"},
		},
	})
	if err != nil {
		log.Printf("Failed to create OpenAI client: %v", err)
	} else {
		fmt.Printf("✅ OpenAI client created: %s\n", openaiClient.Provider().Name())
		openaiClient.Close()
	}

	// Anthropic client (won't work without real API key)
	anthropicClient, err := omnillm.NewClient(omnillm.ClientConfig{
		Providers: []omnillm.ProviderConfig{
			{Provider: omnillm.ProviderNameAnthropic, APIKey: "demo-anthropic-key"},
		},
	})
	if err != nil {
		log.Printf("Failed to create Anthropic client: %v", err)
	} else {
		fmt.Printf("✅ Anthropic client created: %s\n", anthropicClient.Provider().Name())
		anthropicClient.Close()
	}

	// Bedrock client (requires AWS credentials)
	bedrockClient, err := omnillm.NewClient(omnillm.ClientConfig{
		Providers: []omnillm.ProviderConfig{
			{Provider: omnillm.ProviderNameBedrock, Region: "us-east-1"},
		},
	})
	if err != nil {
		log.Printf("⚠️  Bedrock client creation failed (expected without AWS creds): %v", err)
	} else {
		fmt.Printf("✅ Bedrock client created: %s\n", bedrockClient.Provider().Name())
		bedrockClient.Close()
	}

	// Ollama client (works locally, no credentials needed)
	ollamaClient, err := omnillm.NewClient(omnillm.ClientConfig{
		Providers: []omnillm.ProviderConfig{
			{Provider: omnillm.ProviderNameOllama, BaseURL: "http://localhost:11434"},
		},
	})
	if err != nil {
		log.Printf("⚠️  Ollama client creation failed (is Ollama running?): %v", err)
	} else {
		fmt.Printf("✅ Ollama client created: %s\n", ollamaClient.Provider().Name())
		ollamaClient.Close()
	}

	fmt.Println()
	fmt.Println("Benefits of this architecture:")
	fmt.Println("1. 🔌 Pluggable: Easy to add new LLM providers")
	fmt.Println("2. 🎯 Unified: Same API for all providers")
	fmt.Println("3. 🧪 Testable: Provider interface can be mocked")
	fmt.Println("4. 📦 Modular: Each provider is self-contained")
	fmt.Println("5. 🔧 Maintainable: Clear separation of concerns")
	fmt.Println("6. 🏠 Local + Cloud: Mix local (Ollama) and cloud providers")

	// Show example request structure
	fmt.Println()
	fmt.Println("Example unified request structure:")
	fmt.Println("Cloud model example:")
	cloudReq := &omnillm.ChatCompletionRequest{
		Model: omnillm.ModelGPT4o, // OpenAI cloud model
		Messages: []omnillm.Message{
			{Role: omnillm.RoleSystem, Content: "You are a helpful assistant."},
			{Role: omnillm.RoleUser, Content: "Hello, world!"},
		},
		MaxTokens:   &[]int{100}[0],
		Temperature: &[]float64{0.7}[0],
	}
	fmt.Printf("  Model: %s (OpenAI)\n", cloudReq.Model)

	fmt.Println()
	fmt.Println("Local model example:")
	localReq := &omnillm.ChatCompletionRequest{
		Model: "llama3", // Ollama local model
		Messages: []omnillm.Message{
			{Role: omnillm.RoleSystem, Content: "You are a helpful assistant."},
			{Role: omnillm.RoleUser, Content: "Hello, world!"},
		},
		MaxTokens:   &[]int{100}[0],
		Temperature: &[]float64{0.7}[0],
	}
	fmt.Printf("  Model: %s (Ollama)\n", localReq.Model)

	fmt.Println()
	fmt.Println("🎉 The same request structure works with ALL providers!")
	fmt.Println("   Switch from cloud to local by just changing the client config!")
}
