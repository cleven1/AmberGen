package agent

import (
  "context"
)

type Chain struct {
  agents    []Agent
  maxRounds int
}

func NewChain(maxRounds int, agents ...Agent) *Chain {
  return &Chain{
    agents:    agents,
    maxRounds: maxRounds,
  }
}

func (c *Chain) Execute(ctx context.Context, input string) (string, error) {
  result := input
  
  for round := 0; round < c.maxRounds; round++ {
    for _, agent := range c.agents {
      var err error
      result, err = agent.Execute(ctx, result)
      if err != nil {
        return "", err
      }
    }
  }
  
  return result, nil
}
