package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")

	if err != nil {
		log.Println("Error loading .env file:", err)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err = LoadConfig()
	if err != nil {
		log.Println("Failed to load config:", err)
		return
	}

	g := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel(cfg.Model),
	)

	messageGeneratorFlow := genkit.DefineFlow(g, "messageGeneratorFlow", func(ctx context.Context, input *GenerateMessage) (*Messsage, error) {
		prompt := fmt.Sprintf(`Generate a message given that you are a user in chatbot app with capabilities to read data from the app. 
		Such as read docs, read tramites and read processes. The message should be based on the following scenario:
Actor: %s
Action: The user wants to ask %s

The message should be a natural language message that a user would send in the chatbot app based on the given scenario. In the language of the actor and action.`, input.Roles[0], input.Action)

		message, _, err := genkit.GenerateData[Messsage](ctx, g,
			ai.WithPrompt(prompt),
		)

		if err != nil {
			log.Println("Error generating message:", err)
			return nil, err
		}

		return message, nil
	})

	evaluatorFlow := genkit.DefineFlow(g, "evaluatorFlow", func(ctx context.Context, scenario *ScenarioDefinition) (*Score, error) {
		prompt := fmt.Sprintf(`You are an evaluator for a chatbot app. Evaluate the following response based on the scenario and message provided.
Scenario:
Actor: %s
Action: The user wants to ask %s

Message:
%s

Response:
%s

Score the response on a scale of 0 to 5, where 0 means the response is completely incorrect and 5 means it is perfect. Provide detailed feedback explaining the score and how the response could be improved in the language of the actor and action.`, scenario.Roles[0], scenario.Action, scenario.Message, scenario.Response)

		score, _, err := genkit.GenerateData[Score](ctx, g,
			ai.WithPrompt(prompt),
		)

		if err != nil {
			log.Println("Error generating score:", err)
			return nil, err
		}

		return score, nil
	})

	reviewerFlow := NewReviewerFlow(ctx, g, messageGeneratorFlow, evaluatorFlow, SendMessage)

	scenarios, err := ReadScenarios()
	if err != nil {
		log.Println("Failed to read scenarios:", err)
		return
	}
	var scores []Score

	for i := 0; i < len(scenarios); i++ {
		if err := ctx.Err(); err != nil {
			log.Println("Execution interrupted, stopping scenarios:", err)
			break
		}

		score, err := reviewerFlow.Run(ctx, &scenarios[i])
		if err != nil {
			if ctx.Err() != nil {
				log.Println("Execution interrupted while running scenario")
				break
			}
			log.Printf("Error running scenario '%s': %v\n", scenarios[i].Action, err)
			continue
		}
		scores = append(scores, *score)
	}

	SaveScores(scores)
	SummarizeScores(scores)
}
