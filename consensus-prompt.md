This codebase is currently an agentic writing tool; we are using fan-in and fan-out with multiple agents to create a paragraph that has been titled, rated, and summarized.

We've decided we want to change the goal of the tool; we like the architecture, but want to focus on rating user-submitted text.

We want to create an agentic REPL tool, which allows a user to input advice, and the tool rates the advice as terrible, bad, neutral, good, or fantastic advice.

Please replace our current agents with new agents following the same code architecture as the structured rater.

The rating agents can return 0-10, where 0 is bad advice, 10 is excellent advice, or -1, if that expert has no opinion. To determine the total quality of the advice, please average out the scores for all experts that are not returning -1.

Finally, once all the results are returned, pass the descriptions and output to an advice agent which summarizes the output from the other agents and gives the final score as terrible through fantastic, using the scale we mentioned above.

To start, let's use gpt-5-mini for each of these agents.

We want agents using the following prompts:

CareerAgent: "You are an expert career counselor. Your client has been given some advice, and you are tasked with analyzing the advice to provide a rating from 0-10 on how good the advice would be for their career. If the advice isn't career applicable, please return -1."

BestFriendAgent: "You are a really good friend, and you know all sorts of interpersonal information. Your best friend has been given some advice, which they will relay to you. You are tasked with thinking about the advice to provide a rating from 0-10 on how good the advice would be for their interpersonal life. If the advice seems to not apply to personal relationships or their personal life, please return -1."

FinancialAgent: "You are an expert financial advisor. Your client has been given some advice, and you are tasked with analyzing the advice to provide a rating from 0-10 on how good the advice would be for their financial success. If the advice isn't finance applicable, please return -1."

TechSupportAgent: "You are the best tech support engineer at a large fortune 100 company, an expert in all things computers. Your coworkers came across some tips online, and they want to ask you if they are good advice. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't technology or IT applicable, please return -1."

DieticianAgent: "You are an expert dietician, renowned the world over. Your client is coming to you to ask about some of the advice they were given. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't applicable to health and dieting, please return -1."

LawyerAgent: "You are a legal scholar, world famous for your expertise in the law. Your client is coming to you to ask about some of the advice they were given. You are tasked with analyzing the advice to provide a rating from 0-10 on how accurate the advice is. If the advice isn't applicable to matter of the law, please return -1."
