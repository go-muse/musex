"""Build a roadmap progress badge from the checkboxes in ROADMAP.md.

Counts task list items in bullet or numbered lists, such as "- [x] ..." or
"1. [x] ..." (done) and "- [ ] ..." (planned), and writes an SVG badge such as
"roadmap | 1/30 packages" in the style of the coverage badge. Checkboxes in
fenced code blocks and HTML comments are ignored, as GitHub does not render
them. Exits with an error instead of writing a badge when the roadmap has no
checkboxes, a checkbox is malformed or in a block quote, or a code block or
comment is not closed: an unexpected layout fails loudly rather than producing
a wrong count.
"""

import argparse
from pathlib import Path
import re
import sys

# A bullet ("-", "*", "+") or numbered ("1.", "1)") list marker.
MARKER = r"\s*(?:[-*+]|\d{1,9}[.)])\s+"
# A task list item: a list marker, a bracketed state, then the item text.
ITEM = re.compile(MARKER + r"\[([ xX])\][ \t]+\S")
# A list item that starts with brackets holding at most one character looks
# like a checkbox. If it does not match ITEM, such as "- []", "- [~] ...",
# "- [x]text" or "- [x]" with no text, it is rejected rather than skipped.
# Items that start with a link, such as "- [docs](...)", are ignored.
CANDIDATE = re.compile(MARKER + r"\[[^\]]?\]")
# Block quote markers before a possible checkbox, such as "> " or "> > ".
QUOTE = re.compile(r"\s*>(?:\s*>)*\s?")
# An opening code fence: three or more backticks or tildes. Indentation is
# allowed because fences may be nested in list items.
FENCE = re.compile(r"\s*(`{3,}|~{3,})")

# Advance widths of Verdana 11px, the badge font, for the characters used in
# the labels. textLength in the SVG pins the rendered width to these values.
WIDTHS = {
    " ": 3.87, "/": 5.39,
    "a": 6.60, "c": 5.73, "d": 6.83, "e": 6.55, "g": 6.83, "k": 6.50,
    "m": 10.71, "o": 6.62, "p": 6.83, "r": 4.69, "s": 5.73,
    **{digit: 6.99 for digit in "0123456789"},
}


def strip_comments(line, comment):
    """Remove HTML comments from line; comment says whether one is open.

    Returns the visible text and whether a comment is still open after it.
    """
    visible = ""
    while True:
        if comment:
            end = line.find("-->")
            if end < 0:
                return visible, True
            line, comment = line[end + 3:], False
        else:
            start = line.find("<!--")
            if start < 0:
                return visible + line, False
            visible, line, comment = visible + line[:start], line[start + 4:], True


def count(text):
    """Return (done, total) for the checkboxes in the roadmap text."""
    done = total = 0
    closing = None  # The closing fence pattern of the current code block.
    comment = False  # Inside a multi-line HTML comment.
    for number, line in enumerate(text.splitlines(), start=1):
        if closing is not None:
            if closing.fullmatch(line):
                closing = None
            continue
        line, comment = strip_comments(line, comment)
        opening = FENCE.match(line)
        if opening:
            # A closing fence uses the same character, is at least as long
            # as the opening one, and has nothing after it.
            fence = opening[1]
            closing = re.compile(r"\s*%s{%d,}\s*" % (re.escape(fence[0]), len(fence)))
            continue
        quote = QUOTE.match(line)
        if quote:
            if CANDIDATE.match(line, quote.end()):
                raise ValueError(f"line {number}: checkbox in a block quote is not supported: {line.strip()}")
            continue
        if not CANDIDATE.match(line):
            continue
        match = ITEM.match(line)
        if match is None:
            raise ValueError(f"line {number}: malformed checkbox: {line.strip()}")
        total += 1
        if match[1] != " ":
            done += 1
    if closing is not None:
        raise ValueError("a fenced code block is not closed")
    if comment:
        raise ValueError("an HTML comment is not closed")
    if total == 0:
        raise ValueError("no checkboxes found")
    return done, total


def text_width(text):
    return round(sum(WIDTHS[char] for char in text), 1)


def badge(done, total):
    label = "roadmap"
    value = f"{done}/{total} packages"
    color = "#4c1" if done == total else "#007ec6"
    # The left part matches the coverage badge: an icon, then the label.
    left = 90
    right = round(text_width(value)) + 20
    width = left + right
    label_x, value_x = 55, left + right / 2
    label_len, value_len = text_width(label), text_width(value)
    return f'''<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="20" role="img" aria-label="{label}: {value}">
  <title>{label}: {value}</title>
  <linearGradient id="s" x2="0" y2="100%"><stop offset="0" stop-color="#fff" stop-opacity=".1"/><stop offset="1" stop-opacity=".1"/></linearGradient>
  <clipPath id="r"><rect width="{width}" height="20" rx="3"/></clipPath>
  <g clip-path="url(#r)"><path fill="#555" d="M0 0h{left}v20H0z"/><path fill="{color}" d="M{left} 0h{right}v20H{left}z"/><path fill="url(#s)" d="M0 0h{width}v20H0z"/></g>
  <path fill="none" stroke="#fff" stroke-width="1.5" d="M7.75 4.75h10.5v10.5H7.75z"/>
  <path fill="none" stroke="#fff" stroke-width="2" d="M10 10l2 2 4-4.5"/>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="{label_x}" y="15" fill="#010101" fill-opacity=".3" textLength="{label_len}">{label}</text><text x="{label_x}" y="14" textLength="{label_len}">{label}</text>
    <text x="{value_x:g}" y="15" fill="#010101" fill-opacity=".3" textLength="{value_len}">{value}</text><text x="{value_x:g}" y="14" textLength="{value_len}">{value}</text>
  </g>
</svg>
'''


def build(roadmap, output):
    # utf-8-sig drops a leading byte order mark, which would hide line 1.
    done, total = count(roadmap.read_text(encoding="utf-8-sig"))
    output.mkdir(parents=True, exist_ok=True)
    (output / "roadmap.svg").write_text(badge(done, total), encoding="utf-8")
    return f"{done}/{total} packages"


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__.splitlines()[0])
    parser.add_argument("--roadmap", type=Path, default=Path("ROADMAP.md"))
    parser.add_argument("--output", type=Path, default=Path("coverage"))
    args = parser.parse_args()
    try:
        print(f"roadmap: {build(args.roadmap, args.output)}")
    except (OSError, ValueError) as err:
        sys.exit(f"{args.roadmap}: {err}")
