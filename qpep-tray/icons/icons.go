//go:generate go2array -nolist -platform windows -var ConfigIconData -package icons config_icon.ico
//go:generate go2array -nolist -platform windows -var ExitIconData -package icons exit_icon.ico
//go:generate go2array -nolist -platform windows -var RefreshIconData -package icons refresh_icon.ico
//go:generate go2array -nolist -platform windows -var MainIconData -package icons main_icon.ico
//go:generate go2array -nolist -platform windows -var MainIconWaiting -package icons main_icon_waiting.ico
//go:generate go2array -nolist -platform windows -var MainIconConnected -package icons main_icon_connected.ico
//go:generate go2array -nolist -platform linux -var ConfigIconData -package icons config_icon.png
//go:generate go2array -nolist -platform linux -var ExitIconData -package icons exit_icon.png
//go:generate go2array -nolist -platform linux -var RefreshIconData -package icons refresh_icon.png
//go:generate go2array -nolist -platform linux -var MainIconData -package icons main_icon.png
//go:generate go2array -nolist -platform linux -var MainIconWaiting -package icons main_icon_waiting.png
//go:generate go2array -nolist -platform linux -var MainIconConnected -package icons main_icon_connected.png

package icons

// No actual code, only for go:generate
