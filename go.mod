module github.com/GoGamesLab/Energy

require github.com/GoGamesLab/Inventory v0.0.0

replace (
	github.com/GoGamesLab/Inventory => ../Inventory
)

go 1.26.1
