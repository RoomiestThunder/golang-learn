#!/usr/bin/env python3
"""Заменяет ведущие группы из 4 пробелов на табы внутри Go-кода (блоки ```go и секции заданий)."""
import re, sys, pathlib

def tab(line):
    m = re.match(r'^( +)', line)
    if not m: return line
    n = len(m.group(1)); return '\t' * (n // 4) + ' ' * (n % 4) + line[n:]

def process_lesson(text):
    out, inside, golike = [], False, False
    for line in text.split('\n'):
        if line.startswith('```'):
            if not inside:
                inside = True
                info = line[3:].strip().split()
                golike = (not info) or info[0] == 'go'
            else:
                inside = False
            out.append(line); continue
        out.append(tab(line) if inside and golike else line)
    return '\n'.join(out)

def process_task(text):
    out, sec = [], None
    for line in text.split('\n'):
        if line.startswith('@@ '):
            sec = line[3:].strip(); out.append(line); continue
        out.append(tab(line) if sec in ('starter', 'tests', 'solution') else line)
    return '\n'.join(out)

root = pathlib.Path(__file__).resolve().parent.parent / 'content'
for p in sorted((root / 'lessons').glob('*.md')):
    s = p.read_text(); n = process_lesson(s)
    if n != s: p.write_text(n)
for p in sorted((root / 'pages').glob('*.md')):
    s = p.read_text(); n = process_lesson(s)
    if n != s: p.write_text(n)
for p in sorted((root / 'tasks').glob('*.md')):
    s = p.read_text(); n = process_task(s)
    if n != s: p.write_text(n)
