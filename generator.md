Build an API, where it gets the github link, optionally auth token (for private repos)
it'll get the git repo files from UI. 
Do post order traversal. 
If it receives file extract following information
1. Dependencies. 
2. global var
3. constants
4. init function (optional) functionality
5. struct (fields and types)
6. struct methods 
7. for each method, parse input params, output params, and functionality
8. For each functionality parse to LLM to understand the worklow step - Logically group muliple lines of code to steps so that we could interpret
9. for each workflow step step identify the step name, step type, and identify the workflow name. what's it's changing
10. for each workflow step step identify the "type", "type details", step functionality name, step description, step dependencies, step step variables/objects it depends and variables/objects it's updated
11. step types could be extenal system, database, logic, function call, 
12. step type details should capture the details like extenal system name, database schema, operation name, or logic name or function call details. 
13. step functionality should give the simple functionalitiy name, 
14. step description should describe the points to remember.  