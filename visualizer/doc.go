// Package visualizer provides UTF-8 ANSI tree and boxed card formatting for gatwd data structures and AST nodes.
//
// Key features include:
//   - AST BitFlag Type Identification: Renders AST node types based on their registered BitFlag
//     domain names rather than internal Go reflection struct names.
//   - Boxed Framing & Borders: Supports curved (╭─╮), heavy (┏━┓), and standard UTF-8/ASCII borders.
//   - Outer Border Colors: Customizable border line colors via --border-color=<color>.
//   - Background Styling: Global background colors via --bg=<color>.
//   - Alternating Line Backgrounds: Enhances list/slice readability with cycling background colors
//     via --alternate-bg=<color1>,<color2>,...
//   - Pipeline Integration: StepVisualize transducer formats stream nodes directly within DSL pipelines.
//   - Graphviz/DOT Export: Generates Graphviz DOT graphs via --dot.
package visualizer
