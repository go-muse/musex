"""Build a standalone coverage badge and report using only Go and Python."""

import argparse
from pathlib import Path
import re
import shutil
import subprocess


def build(profile, output):
    output.mkdir(parents=True, exist_ok=True)
    label = "n/a"
    color = "#9f9f9f"
    summary = "No executable Go statements to measure.\n"
    if profile is not None:
        # A header-only profile is valid for packages with no executable code.
        lines = profile.read_text().splitlines()
        if not lines or lines[0] not in ("mode: set", "mode: count", "mode: atomic"):
            raise ValueError("Invalid coverage profile header")
        if any(line.strip() for line in lines[1:]):
            summary = subprocess.check_output(
                ["go", "tool", "cover", f"-func={profile}"], text=True
            )
            match = re.search(r"^total:\s+\(statements\)\s+(\d+(?:\.\d+)?)%$", summary, re.M)
            if match is None or not 0 <= float(match[1]) <= 100:
                raise ValueError("Missing or invalid coverage total")
            label = match[1] + "%"
            value = float(match[1])
            color = "#4c1" if value >= 80 else "#dfb317" if value >= 50 else "#e05d44"
            subprocess.run(
                ["go", "tool", "cover", f"-html={profile}", f"-o={output / 'index.html'}"],
                check=True,
            )
        shutil.copyfile(profile, output / "coverage.out")
    if label == "n/a":
        (output / "index.html").write_text(
            '<!doctype html><html lang="en"><meta charset="utf-8">'
            '<meta name="viewport" content="width=device-width, initial-scale=1">'
            '<title>musex coverage</title><h1>musex coverage: n/a</h1>'
            '<p>No executable Go statements to measure yet.</p></html>\n'
        )
    (output / "coverage.txt").write_text(summary)
    (output / ".nojekyll").touch()
    (output / "coverage.svg").write_text(f'''<svg xmlns="http://www.w3.org/2000/svg" width="144" height="20" role="img" aria-label="coverage: {label}">
  <title>coverage: {label}</title>
  <linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#fff" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>
  <clipPath id="r"><rect width="144" height="20" rx="3"/></clipPath>
  <g clip-path="url(#r)"><path fill="#555" d="M0 0h90v20H0z"/><path fill="{color}" d="M90 0h54v20H90z"/><path fill="url(#s)" d="M0 0h144v20H0z"/></g>
  <path fill="#fff" d="M7 11h3v5H7zM12 7h3v9h-3zM17 4h3v12h-3z"/>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="55" y="15" fill="#010101" fill-opacity=".3">coverage</text><text x="55" y="14">coverage</text>
    <text x="117" y="15" fill="#010101" fill-opacity=".3">{label}</text><text x="117" y="14">{label}</text>
  </g>
</svg>\n''')
    return summary


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    source = parser.add_mutually_exclusive_group(required=True)
    source.add_argument("--profile", type=Path)
    source.add_argument("--no-packages", action="store_true")
    parser.add_argument("--output", type=Path, default=Path("coverage"))
    args = parser.parse_args()
    print(build(args.profile, args.output), end="")
