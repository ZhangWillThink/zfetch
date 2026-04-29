package display

import (
	"os"
	"strings"

	"github.com/WillZhang/zfetch/internal/sysinfo"
)

var logoMap = map[string][]string{
	"ubuntu": {
		"            .-/+oossssoo+/-.",
		"        ´:+ssssssssssssssssss+:",
		"      -+ssssssssssssssssssssssss+-",
		"    /osssssssssssssssssssssssssssso/",
		"  .osssssssssssssssssssssssssssssssso.",
		" ./ssssssssssssssssssssssssssssssssssss/.",
		":ssssssssssssssssssssssssssssssssssssssss:",
		"ossssssssssssssssssssssssssssssssssssssssso",
		"ossssssssssssssssssssssssssssssssssssssssso",
		":ssssssssssssssssssssssssssssssssssssssss:",
		" ./ssssssssssssssssssssssssssssssssssss/.",
		"  .osssssssssssssssssssssssssssssssso.",
		"    /osssssssssssssssssssssssssssso/",
		"      -+ssssssssssssssssssssssss+-",
		"        ´:+ssssssssssssssssss+:",
		"            .-/+oossssoo+/-.",
	},
	"debian": {
		"        _,g$$$$$$$$$$$$$$$,g_",
		"     ,g$$$$$$$$$$$$$$$$$$$$$$$$g,",
		"   ,g$$$$P" + `"` + "     " + `"` + "Y$$,.    `$$$$$g,",
		"  g$$$$  `-.           `  .g$$$$$$$",
		" d$$$$$                 ,$$$$$$$$$$L",
		"d$$$$$$$                $$$$$$$$$$$$L",
		"$$$$$$$$$L             ,$$$$$$$$$$$$$$",
		"$$$$$$$$$$$L          ,d$$$$$$$$$$$$$$$",
		"$$$$$$$$$$$$$,      ,g$$$$$$$$$$$$$$$$$",
		"`$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$´",
		" `$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$P",
		"   `$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$$´",
		"     `Y$$$$$$$$$$$$$$$$$$$$$$$$$$P´",
		"        `" + `"` + "Y$$$$$$$$$$$$$$$$P" + `"` + "´",
		"            `" + `"` + "Y$$$$$$$P" + `"` + "´",
	},
	"arch": {
		"                   /\\",
		"                  /  \\",
		"                 /    \\",
		"                /  __  \\",
		"               /  /  \\  \\",
		"              /  /`''´\\  \\",
		"             /  /      \\  \\",
		"            /  /        \\  \\",
		"           /\\  \\        /  /\\",
		"          /  \\  \\      /  /  \\",
		"         /    \\  \\    /  /    \\",
		"        /      \\  \\__/  /      \\",
		"       /       /`      ´\\       \\",
		"      /       /          \\       \\",
		"     /       /            \\       \\",
		"    /       /              \\       \\",
		"   /_______/                \\_______\\",
	},
	"fedora": {
		"             .´;::::::;´.",
		"         .´:cccccccccccccc:;.",
		"      .:cccccccccccccccccccccc:.",
		"    .:cccccccccccccccccccccccccc:.",
		"  .;cccccccccccccc;.:dddl:.;ccccccc;.",
		" .:cccccccccccccc;OWMKOOXMWd;ccccccc:.",
		".:ccccccccccccccc;KMMc;cc;xMMc;ccccc:.",
		",cccccccccccccccc;MMM.;cc;;WW;cccccc;.",
		":cccccccccccccccc;MMM.;cccccccccccccc:",
		":cccccccccccccccc;MMM0OOO0MMMMMMMMMMM:",
		"ccccccccccccccccc;MMWxdddxXMMMMMMMMM;",
		"ccccccccccccccccc;MMM.;cccccccccccccc;",
		"ccccccccccccccccc;MMW.;ccccccccccccc;",
		"ccccccccccccccccc;MMW.;ccccccccccccc;",
		"ccccccccccccccccc;0MWXXXXNW0;ccccccc:",
		" ccccccccccccccccc;.:odl:.;ccccccccc:",
		"  .:ccccccccccccccccccccccccccccc:;.",
		"    ´.::cccccccccccccccccccc::;.´",
	},
	"centos": {
		"                  ..",
		"                .PLTJ.",
		"               <><><><>",
		"      KKSSV´ 4KKK LJ KKKL.´VSSKK",
		"      KKV´ 4KKKKK LJ KKKKAL ´VKK",
		"      V´ ´ ´VKKKK LJ KKKKV´ ´ ´V",
		"      .4MA.´ ´VKK LJ KKV´ ´.4Mb.",
		"    . KKKKKA.´ ´V LJ V´ ´.4KKKKK .",
		"  .4D KKKKKKKA.´´ LJ ´´.4KKKKKKK FA.",
		" <QDD ++++++++++++  ++++++++++++ GFD>",
		"  ´VD KKKKKKKK´.. LJ ..´KKKKKKKK FV",
		"    ´ VKKKKK´. .4 LJ K. .´KKKKKV ´",
		"       ´VK´. .4KK LJ KKA. .´KV´",
		"     A. . .4KKKK LJ KKKKA. . .4",
		"     KKA. ´KKKKK LJ KKKKK´ .4KK",
		"     KSSA. VKKKK LJ KKKKV .4SSK",
	},
	"opensuse": {
		"           .;ldkO0000Okdl;.",
		"       .;d00xl:^´´´´´´^:ok00d;.",
		"     .d00l´                ´o00d.",
		"   .d0K^´  Okxoc:;,.          ^O0d.",
		"  .0KKl.,0d;lxO0KKKOo:.       lKK0.",
		" ,0KKKO:´;dKWMMMMMWWWWNKx;.  :OKKK0,",
		".OKKKKKKx;.;xKWMMMMMMMMMMW0l,;dKKKKK0.",
		"0KKKKKKKKKd,.;dKWMMMMMMMMW0l,oKKKKKKK0:",
		"lKKKKKKKKKKKd,.:OWMMMMMMW0l,dKKKKKKKKKKl",
		"dKKKKKKKKKKKKKOc,l0XNWNKkc,cOKKKKKKKKKKKKd",
		"dKKKKKKKKKKKKKKKd´.....´cOKKKKKKKKKKKKKKd",
		"lKKKKKKKKKKKKKKKx´  .;xKKKKKKKKKKKKKKKl",
		"oKKKKKKKKKKKKKKk´  ;kKKKKKKKKKKKKKKko",
		" lKKKKKkoolll:´    ´cloooxKKKKKKd´",
		"  cKKKd´               ´dKKKx.",
		"   c0x´                 ´x0x.",
	},
	"redhat": {
		"              \\`-..........-´/",
		"              `.,,:;:::::;,.´",
		"           '-.,:cccccccccccc:;,.-",
		"        '-.,:cccccccccccccccccc:;,.-",
		"      '-.,:cccccccccccccccccccccc:;,.-",
		"     '-.:cccccccccccccccccccccccccc:,.`",
		"     '-.:cccccccccccccccccccccccccc:,.-",
		"      '-.:ccccccccccccccccccccccc:;,.-",
		"        '-,:cccccccccccccccccc:;,.-",
		"           '-.,:cccccccccccc:;,.-",
		"              '-.,:cccccc:;,.-",
		"                 '-.,::;,.-",
		"                   '-..-'",
	},
	"gentoo": {
		"        `.-/+oo+:--.`",
		"     .:+oooooooooooo+:.",
		"   ./oooooooooooooooooo/.",
		"  -oooooooooooooooooooooo-",
		"  +oooooooooooooooooooooo+",
		" ./oooooooooooooooooooooo/.",
		" .+oooooooooooooooooooooo+.",
		"  +oooooooooooooooooooooo+",
		"  -oooooooooooooooooooooo-",
		"   ./oooooooooooooooooo/.",
		"     .:+oooooooooooo+:.",
		"        `.-/+oo+:--.`",
	},
	"nixos": {
		"          :::::::::",
		"        :::::::::::::::",
		"      ::::::::::::::::::::",
		"    :::::::::::`    :::::::::::",
		"   ::::::::::`      `::::::::::",
		"  :::::::::::        :::::::::::",
		" :::::::::::          :::::::::::",
		" ::::::::::`          `::::::::::",
		" :::::::::::          :::::::::::",
		" `:::::::::::.      .:::::::::::",
		"  `::::::::::::::::::::::::::::",
		"   `::::::::::::::::::::::::::",
		"     `::::::::::::::::::::::",
		"       `::::::::::::::::::",
		"          `:::::::::::`",
	},
	"linux": {
		"        .:+++/-.`",
		"     .+ooooooooo+:",
		"    /oooooooooooooo:",
		"  `+oooooooooooooooo+`",
		"  /oooooooooooooooooooo:",
		" .oooooooooooooooooooooo.",
		" -ooooooooooooooooooooooo-",
		" -ooooooooooooooooooooooo-",
		" /oooooooooooooooooooooooo:",
		" `+oooooooooooooooooooooo+`",
		"  /oooooooooooooooooooooo:",
		"   .+oooooooooooooooooo+.",
		"     .:+oooooooooooo+:.",
		"        `.://+++/:.`",
	},
	"alpine": {
		"          .:-:.",
		"        .:/+o+/:`",
		"      `:+ooooooo+:`",
		"     .+ooooooooooo+.",
		"    -+ooooooooooooooo+-",
		"   /ooooooooooooooooooo/",
		"  /ooooooooooooooooooooo/",
		" /ooooooooooooooooooooooo/",
		"/ooooooooooooooooooooooooo/",
		"\\ooooooooooooooooooooooooo\\",
		" \\ooooooooooooooooooooooo/",
		"  \\ooooooooooooooooooooo/",
		"   \\ooooooooooooooooooo/",
		"    `-+oooooooooooo+-`",
		"        `.:++++:.",
	},
	"default": {
		"      ██████╗ ███████╗███████╗████████╗ ██████╗██╗  ██╗",
		"      ╚════██╗██╔════╝██╔════╝╚══██╔══╝██╔════╝██║  ██║",
		"       █████╔╝█████╗  █████╗     ██║   ██║     ███████║",
		"      ██╔═══╝ ██╔══╝  ██╔══╝     ██║   ██║     ██╔══██║",
		"      ██████╗ ██║     ███████╗   ██║   ╚██████╗██║  ██║",
		"      ╚═════╝ ╚═╝     ╚══════╝   ╚═╝    ╚═════╝╚═╝  ╚═╝",
	},
}

func GetLogo(name string) []string {
	if logo, ok := logoMap[name]; ok {
		return logo
	}
	return logoMap["default"]
}

func detectOSLogo() string {
	info := sysinfo.GetOS()
	id := strings.ToLower(info.ID)

	if _, ok := logoMap[id]; ok {
		return id
	}

	name := strings.ToLower(info.Name)
	for key := range logoMap {
		if key != "default" && strings.Contains(name, key) {
			return key
		}
	}

	if _, err := os.Stat("/etc/arch-release"); err == nil {
		return "arch"
	}
	if _, err := os.Stat("/etc/fedora-release"); err == nil {
		return "fedora"
	}
	if _, err := os.Stat("/etc/redhat-release"); err == nil {
		return "redhat"
	}
	if _, err := os.Stat("/etc/debian_version"); err == nil {
		return "debian"
	}
	if _, err := os.Stat("/etc/SuSE-release"); err == nil {
		return "opensuse"
	}
	if _, err := os.Stat("/etc/gentoo-release"); err == nil {
		return "gentoo"
	}
	if _, err := os.Stat("/etc/nix"); err == nil {
		return "nixos"
	}

	return "linux"
}

func ListLogos() []string {
	logos := make([]string, 0, len(logoMap))
	for name := range logoMap {
		logos = append(logos, name)
	}
	return logos
}
