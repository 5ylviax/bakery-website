#!/usr/bin/env python3
"""Install release polling for an existing baking.service (public GitHub repo).

Install: sudo python3 setup-baking-cd.py OWNER/REPO
Deploy:  sudo systemctl start baking-pull.service
Inspect: sudo journalctl -u baking-pull.service -n 40 --no-pager
Enable after the first successful deployment:
         sudo systemctl enable --now baking-pull.timer

Requires a healthy, non-root baking.service, localhost:4001, and releases
tagged deploy-N-N containing baking.tar.gz with baking-server/static/templates.
The package must include every file the app needs; persistent app data must
already use paths outside downloaded releases. No GitHub token is stored.
"""

import fcntl
import hashlib
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
import time
import urllib.error
import urllib.request
from pathlib import Path, PurePosixPath

APP = "baking.service"
JOB = "baking-pull.service"
TIMER = "baking-pull.timer"
BASE = Path("/srv/baking-pull")
CONFIG = Path("/etc/baking-pull.json")
PROGRAM = Path("/usr/local/sbin/baking-pull.py")
UNITS = Path("/etc/systemd/system")
DROP_IN = UNITS / (APP + ".d") / "90-release-pull.conf"
ARTIFACT = "baking.tar.gz"
HEALTH_URL = "http://127.0.0.1:4001/"


def systemctl(*args):
    return subprocess.run(["systemctl", *args], check=True, text=True)


def property_value(name):
    return subprocess.check_output(
        ["systemctl", "show", APP, "-p", name, "--value"], text=True
    ).strip()


def healthy():
    for _ in range(10):
        try:
            with urllib.request.urlopen(HEALTH_URL, timeout=2) as response:
                if response.status == 200 and property_value("SubState") == "running":
                    return True
        except OSError:
            pass
        time.sleep(1)
    return False


def get_release(repo):
    endpoint = f"https://api.github.com/repos/{repo}/releases/latest"
    request = urllib.request.Request(
        endpoint,
        headers={"Accept": "application/vnd.github+json", "User-Agent": "baking-pull"},
    )
    try:
        with urllib.request.urlopen(request, timeout=20) as response:
            release = json.load(response)
    except urllib.error.HTTPError as error:
        if error.code == 404:
            raise ValueError(
                f"No public release found for {repo}. Check OWNER/REPO and GitHub Releases."
            ) from error
        raise
    tag = release["tag_name"]
    if not re.fullmatch(r"deploy-[0-9]+-[0-9]+", tag):
        raise ValueError("Latest GitHub release is missing the expected deploy tag.")
    for asset in release["assets"]:
        if asset["name"] == ARTIFACT:
            url = f"https://github.com/{repo}/releases/download/{tag}/{ARTIFACT}"
            digest = asset.get("digest") or ""
            if re.fullmatch(r"sha256:[a-f0-9]{64}", digest):
                return tag, url, digest.removeprefix("sha256:")
            raise ValueError("The release asset has no SHA-256 digest yet.")
    raise ValueError(f"Latest GitHub release does not contain {ARTIFACT}.")


def unpack(package, stage):
    with tarfile.open(package, "r:gz") as archive:
        for item in archive.getmembers():
            path = PurePosixPath(item.name)
            if (
                path.is_absolute()
                or not path.parts
                or ".." in path.parts
                or path.parts[0] not in {"baking-server", "static", "templates"}
                or not (item.isdir() or item.isfile())
            ):
                raise ValueError(f"Unexpected archive entry: {item.name!r}")
        archive.extractall(stage, filter="data")
    if (
        not (stage / "baking-server").is_file()
        or not os.access(stage / "baking-server", os.X_OK)
        or not all((stage / name).is_dir() for name in ("static", "templates"))
    ):
        raise ValueError("Package must contain executable baking-server, static/, and templates/.")


def switch(target):
    temporary = BASE / "next"
    temporary.unlink(missing_ok=True)
    temporary.symlink_to(target)
    temporary.replace(BASE / "current")


def prepare_release(release, url, expected_digest, env_file):
    digest_file = release / ".release-sha256"
    if digest_file.is_file() and digest_file.read_text().strip() == expected_digest:
        return
    if release.exists():
        raise ValueError(f"{release} exists with an unexpected digest. Inspect it before retrying.")
    with tempfile.TemporaryDirectory(dir=BASE) as work:
        package = Path(work) / ARTIFACT
        checksum = hashlib.sha256()
        with urllib.request.urlopen(url, timeout=30) as response, package.open("wb") as output:
            for chunk in iter(lambda: response.read(1024 * 1024), b""):
                checksum.update(chunk)
                output.write(chunk)
        if checksum.hexdigest() != expected_digest:
            raise ValueError("Release checksum mismatch. The live site was left alone.")
        stage = Path(work) / "site"
        stage.mkdir()
        stage.chmod(0o755)
        unpack(package, stage)
        if env_file.is_file():
            (stage / ".env").symlink_to(env_file)
        (stage / digest_file.name).write_text(expected_digest + "\n")
        stage.rename(release)


def restore(previous):
    switch(previous)
    systemctl("restart", APP)
    if not healthy():
        raise RuntimeError("The previous release failed its health check during rollback.")


def deploy():
    pending = BASE / ".previous-release"
    if pending.is_symlink():
        restore(pending.resolve(strict=True))
        pending.unlink()
        print("Restored the previous release after an interrupted deployment.")

    settings = json.loads(CONFIG.read_text())
    tag, url, expected_digest = get_release(settings["repo"])
    current = (BASE / "current").resolve(strict=True)
    release = BASE / tag
    rejected = BASE / (tag + ".rejected")
    if current == release:
        print(f"{tag} is already installed.")
        return
    if rejected.exists():
        print(f"{tag} failed its health check previously. Waiting for a new release.")
        return

    env_file = Path(settings["original_directory"]) / ".env"
    prepare_release(release, url, expected_digest, env_file)
    pending.symlink_to(current)
    try:
        switch(release)
        systemctl("restart", APP)
        if not healthy():
            raise RuntimeError("New release failed the local health check.")
    except (OSError, subprocess.CalledProcessError, RuntimeError):
        restore(current)
        rejected.touch()
        pending.unlink()
        print(f"Restored the previous release after {tag} failed.", file=sys.stderr)
        raise
    pending.unlink()
    print(f"Installed {tag}. The site returned HTTP 200.")


def service_argv():
    return re.findall(r"argv\[\]=(.*?)\s*;\s*ignore_errors=", property_value("ExecStart"))


def install(repo):
    if not re.fullmatch(r"[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+", repo):
        raise ValueError("Supply the GitHub repository as OWNER/REPO.")
    for path in (BASE, CONFIG, PROGRAM, DROP_IN, UNITS / JOB, UNITS / TIMER):
        if os.path.lexists(path):
            raise ValueError(f"{path} already exists. Existing deployment setup was preserved.")

    user = property_value("User")
    workdir = Path(property_value("WorkingDirectory"))
    binary = workdir / "baking-server"
    argv = service_argv()
    if (
        user in ("", "0", "root")
        or not workdir.is_absolute()
        or not binary.is_file()
        or len(argv) != 1
        or re.search(r"[\s%\\]", str(workdir))
    ):
        raise ValueError(
            "Expected non-root baking.service running baking-server from its working directory. "
            "Run: systemctl cat baking.service"
        )
    command = argv[0].strip()
    allowed_commands = (
        str(binary),
        f"{binary} -addr=127.0.0.1:4001",
        f"{binary} -addr 127.0.0.1:4001",
    )
    if command not in allowed_commands:
        raise ValueError("Unexpected baking.service startup arguments. Run: systemctl cat baking.service")
    new_command = str(BASE / "current/baking-server") + command[len(str(binary)):]
    if not healthy():
        raise RuntimeError("Existing baking.service must be running and returning HTTP 200 on port 4001.")
    tag, _, _ = get_release(repo)
    print(f"Found GitHub release {tag}.")

    job_text = (
        "[Unit]\nDescription=Install the latest verified Baking release\n"
        "Wants=network-online.target\nAfter=network-online.target\n"
        f"\n[Service]\nType=oneshot\nExecStart=/usr/bin/python3 {PROGRAM} --run\n"
        "UMask=0022\nTimeoutStartSec=5min\n"
    )
    timer_text = (
        "[Unit]\nDescription=Check for Baking releases every two minutes\n"
        "\n[Timer]\nOnActiveSec=2min\nOnUnitInactiveSec=2min\nAccuracySec=5s\n"
        "\n[Install]\nWantedBy=timers.target\n"
    )
    original_directory = workdir.resolve(strict=True)
    files = {
        CONFIG: json.dumps({"repo": repo, "original_directory": str(original_directory)}) + "\n",
        PROGRAM: Path(__file__).read_text(),
        UNITS / JOB: job_text,
        UNITS / TIMER: timer_text,
        DROP_IN: (
            f"[Service]\nWorkingDirectory={BASE}/current\nExecStart=\n"
            f"ExecStart={new_command}\n"
        ),
    }
    created = []
    BASE.mkdir(mode=0o755)
    try:
        BASE.chmod(0o755)
        (BASE / "current").symlink_to(original_directory)
        DROP_IN.parent.mkdir(mode=0o755, parents=True, exist_ok=True)
        for path, content in files.items():
            with path.open("x") as output:
                created.append(path)
                output.write(content)
            path.chmod(0o644)
        systemctl("daemon-reload")
        if (
            property_value("WorkingDirectory") != str(BASE / "current")
            or service_argv() != [new_command]
        ):
            raise RuntimeError("Another systemd drop-in overrides the release settings.")
    except (OSError, RuntimeError, subprocess.CalledProcessError):
        for path in reversed(created):
            path.unlink()
        shutil.rmtree(BASE)
        systemctl("daemon-reload")
        raise
    print("Installed deployment files. Run the first deployment, then enable the timer:")
    print(f"  sudo systemctl start {JOB}")
    print(f"  sudo journalctl -u {JOB} -n 40 --no-pager")
    print(f"  sudo systemctl enable --now {TIMER}")


def main():
    if os.geteuid() != 0:
        raise ValueError("Run this script with sudo.")
    if not hasattr(tarfile, "data_filter"):
        raise ValueError("Update Ubuntu's Python packages to get secure tar extraction support.")
    if len(sys.argv) != 2:
        raise ValueError("Usage: sudo python3 setup-baking-cd.py OWNER/REPO")
    if sys.argv[1] == "--run":
        # Serializes direct invocations with the systemd deployment service.
        with (BASE / ".deploy.lock").open("a") as lock:
            fcntl.flock(lock, fcntl.LOCK_EX)
            deploy()
    else:
        install(sys.argv[1])


if __name__ == "__main__":
    try:
        main()
    except (ValueError, OSError, KeyError, RuntimeError, tarfile.TarError, subprocess.CalledProcessError) as error:
        sys.exit(str(error))
