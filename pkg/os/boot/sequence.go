package boot

// BootResult is a boot: its stages on a timeline (in milliseconds) and what
// each one does.
type BootResult struct {
	Stages []Stage  `json:"stages"`
	Notes  []string `json:"notes"`
	Total  int      `json:"total"`
}

// Sequence lays out a PC booting, with uefi choosing UEFI (with GPT and an
// EFI system partition) or legacy BIOS (with a 512-byte MBR boot sector),
// and the number of system services init starts in parallel:
//
//	power on  the CPU starts at the reset vector in firmware
//	POST      the firmware tests and initializes memory and devices
//	loader    BIOS runs the MBR's 446 bytes of code, which loads more;
//	          UEFI reads the EFI partition and runs a boot manager directly
//	kernel    decompressed into memory, it sets up paging, drivers and
//	          mounts an initial RAM disk, then the real root file system
//	init      PID 1 (systemd) starts services, in parallel where the
//	          dependencies allow
//	login     a login prompt or display manager
//
// The durations are typical, not measured.
//
// Time: O(services). Space: O(services).
func Sequence(uefi bool, services int) (BootResult, error) {
	if services < 1 || services > 12 {
		return BootResult{}, inputErr("1-12 services")
	}
	var res BootResult
	t := 0
	add := func(lane, what string, d int, note string) {
		res.Stages = append(res.Stages, Stage{Lane: lane, What: what, Start: t, End: t + d, State: "running"})
		res.Notes = append(res.Notes, note)
		t += d
	}
	add("firmware", "power on, reset vector", 50, "the CPU starts executing firmware at a fixed address; nothing else is set up")
	add("firmware", "POST", 800, "power-on self-test: check and initialize RAM, the CPU, and find the devices")
	if uefi {
		add("firmware", "UEFI boot manager", 200, "UEFI reads the GPT, finds the EFI system partition (FAT32) and runs the boot loader file from it")
		add("loader", "boot loader (systemd-boot / GRUB)", 300, "the loader shows its menu, then loads the kernel and the initramfs into memory")
	} else {
		add("firmware", "read the MBR", 100, "BIOS loads sector 0 (512 bytes: 446 of code, the partition table, the 0x55AA signature) and jumps to it")
		add("loader", "stage 1 and 2 loader (GRUB)", 400, "446 bytes are too few, so the boot code loads a second stage, which can read file systems")
	}
	add("kernel", "decompress, set up paging", 300, "the kernel unpacks itself, switches to 64-bit mode with paging, and starts the scheduler")
	add("kernel", "drivers, initramfs, mount root", 900, "it probes hardware, runs the initial RAM disk to find the root file system, and mounts it")
	start := t
	longest := 0
	for i := 0; i < services; i++ {
		d := 200 + 150*(i%4)
		res.Stages = append(res.Stages, Stage{Lane: "init", What: serviceNames[i%len(serviceNames)], Start: start, End: start + d, State: "running"})
		longest = max(longest, d)
	}
	res.Notes = append(res.Notes, "init (PID 1) starts the services in parallel, so this stage takes as long as the slowest, not the sum")
	t += longest
	add("login", "login prompt", 100, "the system is up")
	res.Total = t
	return res, nil
}

var serviceNames = []string{"udev", "network", "journald", "ssh", "cron", "dbus", "time sync", "display manager", "bluetooth", "printing", "firewall", "logind"}
