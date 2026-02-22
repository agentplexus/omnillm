# Tool Calling

OmniLLM supports function/tool calling for building agentic workflows. Tools allow the LLM to request specific actions that your application can execute.

## Basic Tool Calling

```go
response, err := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: omnillm.ModelGPT4o,
    Messages: []omnillm.Message{
        {Role: omnillm.RoleUser, Content: "What's the weather in Tokyo?"},
    },
    Tools: []omnillm.Tool{
        {
            Type: "function",
            Function: omnillm.ToolFunction{
                Name:        "get_weather",
                Description: "Get current weather for a location",
                Parameters: map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "location": map[string]any{
                            "type":        "string",
                            "description": "City name",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
})
```

## Handling Tool Calls

When the LLM wants to call a tool, the response includes tool calls:

```go
if len(response.Choices) > 0 && len(response.Choices[0].Message.ToolCalls) > 0 {
    for _, toolCall := range response.Choices[0].Message.ToolCalls {
        if toolCall.Function.Name == "get_weather" {
            // Parse arguments
            var args struct {
                Location string `json:"location"`
            }
            json.Unmarshal([]byte(toolCall.Function.Arguments), &args)

            // Execute the tool
            weather := getWeather(args.Location)

            // Send tool result back to LLM
            response, err = client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
                Model: omnillm.ModelGPT4o,
                Messages: []omnillm.Message{
                    {Role: omnillm.RoleUser, Content: "What's the weather in Tokyo?"},
                    response.Choices[0].Message, // Assistant message with tool call
                    {
                        Role:       omnillm.RoleTool,
                        Content:    weather,
                        ToolCallID: &toolCall.ID,
                    },
                },
            })
        }
    }
}
```

## Provider Support

| Provider | Tool Calling |
|----------|--------------|
| OpenAI | Yes |
| Anthropic | Yes |
| X.AI (Grok) | Yes |
| Google Gemini | Partial |
| Ollama | Model-dependent |

## Tool Types

```go
type Tool struct {
    Type     string       `json:"type"`     // "function"
    Function ToolFunction `json:"function"`
}

type ToolFunction struct {
    Name        string         `json:"name"`
    Description string         `json:"description"`
    Parameters  map[string]any `json:"parameters"` // JSON Schema
}

type ToolCall struct {
    ID       string           `json:"id"`
    Type     string           `json:"type"`
    Function ToolCallFunction `json:"function"`
}

type ToolCallFunction struct {
    Name      string `json:"name"`
    Arguments string `json:"arguments"` // JSON string
}
```
