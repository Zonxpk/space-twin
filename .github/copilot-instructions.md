# Copilot Instructions: UI/UX Pro Max

When user asks for UI/UX design, landing pages, style guides, or web interfaces:

1. **Use the UI/UX Pro Max tool** to generate a complete design system first.
   - Command: `py .github/prompts/ui-ux-pro-max/scripts/search.py "<query>" --design-system -p "<ProjectName>"`
   - Example: `py .github/prompts/ui-ux-pro-max/scripts/search.py "AI Floorplan SaaS" --design-system -p "Floorplan Whiteboard"`

2. **For specific components or guidelines**:
   - Use `--domain` flag (style, ux, typography, etc.)
   - Command: `py .github/prompts/ui-ux-pro-max/scripts/search.py "<query>" --domain <domain>`

3. **For implementation details**:
   - Use `--stack` flag (default: html-tailwind, or react, nextjs, vue, etc.)
   - Command: `py .github/prompts/ui-ux-pro-max/scripts/search.py "<query>" --stack <stack>`

4. **Always follow the generated design system** when writing code.

Do not halluciante styles. Use the tool to retrieve professional guidelines.
