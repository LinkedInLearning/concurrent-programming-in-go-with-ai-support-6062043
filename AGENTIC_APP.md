# Agentic Storywriter
Now, we need to create a new application for our demo. We're going to create an agentic, adventure novel storywriter.

We need this application to follow a supervisor agent, agentee structure.
Goal: User submits a couple of sentences for a story they want written. Supervisor takes over from here, and calls all other sub-agents in a loop.
Each of the subagents are running persistently as background goroutines. Communication with these agents occurs over channels.

Many of these operations are independent, so please parallelize them as much as possible.
Ensure there is a mechanism in place for memory pressure; there is an agent described below which can compact sessions for the supervisor. Use tiktoken (for tiktoken, use the 4o counter) to determine when we are at 70% of the context threshold and compact the session before moving on.

Use the correct message types for subagent calling, use context7 and look at the openai documentation. tool calls might be the correct move, but I'll leave that up to you.

We need a form of persistence. Have the program create a workspace folder, and store all the output information in there. Need to store the complete results from each of the subagent calls. 

Agents also need tools to edit files, so they can write their stories. LOCK these tool calls to just the workspace folder. They should be allowed to edit, create, and delete .md files only.

Here are the subagents:

- Plot designer: This agent is in charge of designing the story plot. Every plot should look like this:

1. A background scene is established, and some of the basic worldbuilding rules are established.
2. The reader's reason for caring is established; some background information on the main protagonist is established as they go about their life.
3. The hero's journey begins: Something the protagonist wants or needs is introduced, and their life is thrown upside down as a result.
4. The hero learns Something Horrible will happen if their goals are not achieved
5. The protagonist fails and suffers, but all is not lost
6. The hero learns from their mistakes, and they emerge from the situation stronger than before
7. The hero learns the villian or obstacle or dreadful thing is more powerful, more difficult, or otherwise even scarier than initially perceived.
8. The hero overcomes anyway. They might barely scrape by, and they should pay a steep price. But a victory takes place nonetheless.
9. The hero returns and adjusts to new life post-plot, and small or large changes to their environment are now evident.

This agent needs to provide an answer to *each* of the above entries at a minimum.

- Worldbuilder agent: This agent needs to look at the plot, and develop a world in which the plot can take place. Can it take place in the real world? Must it be a fairytale world? Does magic exist? fireballs and dragons? What time period is it set in? Is it an alternate history? The goal is to be creative and come up with something special to enhance the plot.

- Plot expander agent: This agent looks at the initial entries from the plot designer, loads in the information from the worldbuilder, and expands each entry in the plot from a couple of sentences into a full paragraph, checking carefully for plot holes and avoiding the worst cliches, like deus ex machina.

- Character developer: This agent is in charge of looking at the plot, the world design, and drawing up several very special characters. The protagonist, a villian if applicable, and any supporting characters. For every character, the agent needs to come up with a relevant backstory, lore, and descriptive characteristics on their physical form. Characters should all have names that fit their world.

- Author agent: This agent makes things come alive. The goal of the author agent is to actually write the story, one chapter at a time. The chapters should be short, no longer than 2 pages each. The author agent should be given the character information, the worldbuilding information, and the plot expander output.

- Story summarizer agent: This agent takes a completed chapter from an author agent and summarizes it down to a single paragraph, such that the editor and supervisor can keep the whole story in memory at once.

- Supervisor summary agent: this agent takes in the full agentic history in memory, and summarizes it down to the initial prompt, a bulleted list of all things completed, and the full text of the most recent interaction.

- Editor agent: This agent needs to run at least once per chapter. Here, the editor is in charge of making sure the story is coherent across the chapters. The editor should be fed in the summaries of every chapter, plus the full current chapter, and can edit the text to make sure there are no plot holes or major contradictions.
