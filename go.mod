module github.com/GoGamesLab/Energy

require (
    github.com/GoGamesLab/Inventory v0.0.0
    github.com/GoGamesLab/Materials v0.0.0
)

replace (
    github.com/GoGamesLab/Inventory => ../Inventory
    github.com/GoGamesLab/Materials => ../Materials
)

go 1.26.1
