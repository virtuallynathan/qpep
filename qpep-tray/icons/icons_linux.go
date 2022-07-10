//go:build linux
// +build linux

//go:generate sh -c "2goarray ConfigIconData icons < config_icon.png > config_icon_linux.go"
//go:generate sh -c "2goarray ExitIconData icons < exit_icon.png > exit_icon_linux.go"
//go:generate sh -c "2goarray RefreshIconData icons < refresh_icon.png > refresh_icon_linux.go"
//go:generate sh -c "2goarray MainIconData icons < main_icon.png > main_icon_linux.go"
//go:generate sh -c "2goarray MainIconWaiting icons < main_icon_waiting.png > main_icon_waiting_linux.go"
//go:generate sh -c "2goarray MainIconConnected icons < main_icon_connected.png > main_icon_connected_linux.go"

package icons

// No actual code, only for go:generate
