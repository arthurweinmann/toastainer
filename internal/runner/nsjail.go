package runner

import (
	"os/exec"
	"strconv"
)

// env items must be of the form key=value
func nsjailCommand(chroot, cwd string, timeoutSec int, env []string, ip, gw, iface, nm string, usercmd ...string) *exec.Cmd {
	// TODO: fok nsjail to control logs and only log execve errors in a way that doesn t transpire that we use nsjail

	cmd := exec.Command(nsjailPAth,
		"-Mo",
		"--chroot", chroot,
		"--rw",
		"--user", "0:1000:1",
		"--group", "0:1000:1",
		"--hostname", "toastate",
		"--cwd", cwd,
		"--rlimit_as", "hard",
		"--rlimit_nofile", "hard",
		"--rlimit_nproc", "hard",
		"--rlimit_fsize", "hard",
		"--rlimit_core", "hard",
		"--rlimit_cpu", "hard",
		"--rlimit_stack", "hard",
		"--seccomp_string", kafelPolicy,
		"--macvlan_iface", iface,
		"--macvlan_vs_ip", ip,
		"--macvlan_vs_nm", nm,
		"--macvlan_vs_gw", gw,
	)

	for i := 0; i < len(env); i++ {
		cmd.Args = append(cmd.Args, "-E", env[i])
	}

	if timeoutSec > 0 {
		cmd.Args = append(cmd.Args, "--time_limit", strconv.Itoa(timeoutSec))
	}

	if DebugMode {
		cmd.Args = append(cmd.Args, "-v")
	} else {
		cmd.Args = append(cmd.Args, "-Q") // , "--log_fd", "0"
	}

	cmd.Args = append(cmd.Args, "--")
	cmd.Args = append(cmd.Args, usercmd...)

	return cmd
}

const kafelPolicy = `
POLICY toasters {
  KILL {
    acct,
    add_key,
    adjtimex,
    bpf,
    clock_adjtime,
    clock_settime,
    create_module,
    delete_module,
    finit_module,
    get_kernel_syms,
    get_mempolicy,
    init_module,
    io_cancel,
    io_destroy,
    io_getevents,
    io_setup,
    io_submit,
    ioperm,
    iopl,
    kcmp,
    keyctl,
    kexec_file_load,
    kexec_load,
    lookup_dcookie,
    mbind,
    migrate_pages,
    modify_ldt,
    mount,
    move_pages,
    name_to_handle_at,
    nfsservctl,
    open_by_handle_at,
    perf_event_open,
    personality,
    pivot_root,
    query_module,
    process_vm_readv,
    process_vm_writev,
    ptrace,
    quotactl,
    reboot,
    remap_file_pages,
    request_key,
    seccomp,
    set_mempolicy,
    set_thread_area,
    setns,
    settimeofday,
    syslog,
    swapon,
    swapoff,
    sysfs,
    umount,
    unshare,
    uselib,
    userfaultfd,
    vmsplice
  }
}
USE toasters DEFAULT ALLOW`
