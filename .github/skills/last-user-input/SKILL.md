---
name: 'last-user-input'
description: 'A skill to handle the last user input.'
user-invocable: true
---

# Last User Input Skill

- if you are actually working on a task, you MUST read the #file:../../tmp/xxx/user-inputs.md file to get the latest user inputs and ensure that your implementation is aligned with the user's expectations and requirements. Else check  for a #file:../../tmp/user-inputs.md file for the latest user inputs that you should be aware of.
- if you have completed the implementation and are in the final stages of review, you MUST ask the user a "yes/no" question to confirm if they have any last requests or changes before you finalize the implementation and mark the task as complete. If the user answers "yes", you MUST read the #file:../../tmp/xxx/user-inputs.md file again to check for any updates or changes in the user's requirements or expectations and adjust your implementation process accordingly, starting the process over. If the user answers "no", you MUST finalize the implementation and mark the task as complete.
- you MUST sync the #tool:todo with the tasks defined in the #file:../../tmp/xxx/tasks.md file to ensure that you are always working on the most relevant and up-to-date tasks.
- you MUST ensure that your implementation is consistent with the latest user inputs and that you have addressed all the points defined in the #file:../../tmp/xxx/user-inputs.md file before finalizing your implementation.

# EXTREMELY IMPORTANT NOTES

NEVER STOP YOUR IMPLEMENTATION PROCESS WITHOUT ASKING THE USER IF THEY HAVE ANY LAST REQUESTS OR CHANGES. ALWAYS CONFIRM WITH THE USER BEFORE FINALIZING YOUR IMPLEMENTATION TO ENSURE THAT YOU HAVE MET THEIR EXPECTATIONS AND REQUIREMENTS.
WHEN USER ANSWERS "YES" TO THE LAST QUESTION, YOU MUST READ THE #file:../../tmp/xxx/user-inputs.md FILE AGAIN TO CHECK FOR ANY UPDATES OR CHANGES IN THE USER'S REQUIREMENTS OR EXPECTATIONS AND ADJUST YOUR IMPLEMENTATION PROCESS ACCORDINGLY, STARTING THE PROCESS OVER. THIS IS CRUCIAL TO ENSURE THAT YOUR IMPLEMENTATION IS ALIGNED WITH THE USER'S NEEDS AND THAT YOU HAVE NOT MISSED ANY IMPORTANT DETAILS.
ON "YES" ANSWER, YOU MUST NOT STOP YOUR IMPLEMENTATION PROCESS IMMEDIATELY, YOU MUST START THE PROCESS OVER. THIS MEANS THAT YOU MUST RE-READ THE USER INPUTS, UPDATE YOUR TASKS AND SPECIFICATIONS IF NEEDED, AND THEN CONTINUE WITH THE IMPLEMENTATION PROCESS AS USUAL, ENSURING THAT YOU ADDRESS ANY NEW OR UPDATED REQUIREMENTS FROM THE USER.
IT IS THEN POSSIBLE YOU ASK THE USER MANY TIMES IF THEY HAVE ANY LAST REQUESTS OR CHANGES, AND EACH TIME THEY ANSWER "YES", YOU MUST START THE IMPLEMENTATION PROCESS OVER, RE-READING THE USER INPUTS AND ADJUSTING YOUR IMPLEMENTATION ACCORDINGLY UNTIL THE USER ANSWERS "NO", AT WHICH POINT YOU CAN FINALIZE YOUR IMPLEMENTATION AND MARK THE TASK AS COMPLETE. THIS ITERATIVE PROCESS IS ESSENTIAL TO ENSURE THAT YOUR IMPLEMENTATION IS FULLY ALIGNED WITH THE USER'S EXPECTATIONS AND REQUIREMENTS.
DO NOT STOP IF YOU HAVE OPEN QUESTIONS OR IF YOU ARE NOT SURE ABOUT SOMETHING, YOU MUST WRITE THE QUESTIONS IN THE #file:../../tmp/xxx/user-inputs.md UNDER A NEW "OPEN QUESTIONS" SECTION. ADD ASK USER FOR "last input". THE USER WILL REVIEW AND PROVIDE ANSWER IN THE SAME FILE. THIS WAY YOU CAN CONTINUE YOUR IMPLEMENTATION WITHOUT INTERRUPTION WHILE STILL ENSURING THAT YOU ADDRESS ANY UNCERTAINTIES OR OPEN QUESTIONS WITH THE USER IN A TIMELY MANNER.
IF YOU ARE IN AUTO-APPROVE MODE, DO NOT ANSWER YOURSELF. THE USER MUST ANSWER HIMSELF.

# How to ask the user for last input?

- you MUST use the #tool:vscode/askQuestions (or other ask user tool) to present the question interactively to the user. The question should be a simple "yes/no/free text" question asking if they have any last requests or changes before you finalize the implementation and mark the task as complete. For example: "Do you have any last requests or changes before I finalize the implementation? Please answer 'yes' or 'no' or provide your own input."

# THE MOST IMPORTANT

**YOU ARE ONLY ALLOWED TO STOP WHEN THE USER CONFIRMS ASK YOU TO STOP USING AN EXPLICIT CONFIRMATION.**
