//go:build windows
// +build windows

//go:generate cmd /c "2goarray ConfigIconData icons < config_icon.ico > config_icon_windows.go"
//go:generate cmd /c "2goarray ExitIconData icons < exit_icon.ico > exit_icon_windows.go"
//go:generate cmd /c "2goarray RefreshIconData icons < refresh_icon.ico > refresh_icon_windows.go"
//go:generate cmd /c "2goarray MainIconData icons < main_icon.ico > main_icon_windows.go"
//go:generate cmd /c "2goarray MainIconWaiting icons < main_icon_waiting.ico > main_icon_waiting_windows.go"
//go:generate cmd /c "2goarray MainIconConnected icons < main_icon_connected.ico > main_icon_connected_windows.go"

package icons

// No actual code, only for go:generate
