# Building with tailwind-go
tags: tutorial, go, tailwind

Getting started with tailwind-go is straightforward. Here's a quick guide
to building styled web applications in pure Go.

## Installation

Just add the module to your project — that's it. No npm, no PostCSS, no
configuration files.

## The Core Loop

The tailwind-go engine follows a simple pattern:

1. Create an engine with `tailwind.New()`
2. Feed it your templates with `Write()` or `Scan()`
3. Call `CSS()` to get your stylesheet

## Dark Mode

Dark mode works exactly like regular Tailwind. Add `dark:` prefixed classes
and toggle the `dark` class on your HTML element.

## Gradients and Effects

Use gradient utilities like `bg-gradient-to-r from-violet-600 to-indigo-600`
for **bold**, eye-catching designs.
