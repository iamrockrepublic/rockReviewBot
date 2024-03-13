### Introduction

#### bot_usage
Search *RockReview* on telegram, find a bot named *RockReview* and play with it.

note: Response might be a bit slow due to different region deployment, lambda service is deployed in Singapore while mysql service is deployed in Shanghai. 

#### directories
* app - logic
  * bot_server - logic for review bot
    * dependency - interface definition
    * dependency_go_mock - mock of interface
    * ... - name explains itself
* cmd - runnable
  * bot_config - helper runnable to interact with telegram api
  * bot_lambda - bot runnable deployable to serverless, use webhook 
  * bot_server - bot runnable deployable to ecs, use getUpdates
  * example - example bot for reference
* util - utilities
  * goutil - golang related utilities
  * oss - oss api
  * persist - mysql & redis
  * xlogger - customized log 

