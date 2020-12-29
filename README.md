# Kardia explorer 

Backend for *[explorer](https://explorer.kardiachain.io)*

## Setup

- Generate *[github_private_token](https://github.com/settings/tokens)* with `repo` permission and set as `GITHUB_TOKEN` variables in your system.
- Run `make all` 
- Checkout `Makefile` for more command and use what you need


## Project structure

|--- api: define API for FE  
|--- cfg: define base configuration  
|--- cmd: all entry point here  
|--- contracts: ERC20 and smc  
|--- deployments: docker-compose and dockerfile for deploy/develop  
|--- features: BDD  
|--- kardia: kardia client implement  
|--- metrics: custom metrics for tracking/monitor  
|--- scripts: scripting for execute  
|--- server: logic/db server  
|--- tools: tools for develop  
|--- types  
|--- utils: collections utilities function


## Changelogs
 
#### Version 1.0.0

- Connect with Kai Mainnet through RPC
- Explorer support RestfulAPI
- 


## Support and contributes 

Feel free to create issues if you have any. PRs are welcome and make sure you follow coding guideline. For any information about this project, please contact us at `hello@kardiachain.io`.
