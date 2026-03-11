// DOMCollector tracks nodes removed by virtual scrolling and merges them
// back into the snapshot so we capture the union of all visible content.
//
// Uses arrays (not Sets) to preserve genuinely duplicated nodes
// (e.g. two retweets with identical outerHTML).

const MAX_COLLECTED_SIZE = 5 * 1024 * 1024; // 5MB cap

export class DOMCollector {
  // parent CSS selector -> array of removed node outerHTML (duplicates allowed)
  private removed: Map<string, string[]> = new Map();
  private totalSize = 0;

  handleMutations(mutations: MutationRecord[]): void {
    for (const mutation of mutations) {
      // When a node is re-added (user scrolled back), remove ONE matching
      // entry from collected — not all of them, since duplicates are distinct items.
      for (const node of Array.from(mutation.addedNodes)) {
        if (node.nodeType !== Node.ELEMENT_NODE) continue;
        const html = (node as Element).outerHTML;
        this.removeOneMatch(html);
      }

      // Collect removed element nodes
      if (!mutation.target || !(mutation.target instanceof Element)) continue;
      const parentSel = this.selectorFor(mutation.target);

      for (const node of Array.from(mutation.removedNodes)) {
        if (node.nodeType !== Node.ELEMENT_NODE) continue;
        const html = (node as Element).outerHTML;
        if (this.totalSize + html.length > MAX_COLLECTED_SIZE) continue;

        let arr = this.removed.get(parentSel);
        if (!arr) {
          arr = [];
          this.removed.set(parentSel, arr);
        }
        arr.push(html);
        this.totalSize += html.length;
      }
    }
  }

  /** Merge collected removed nodes into an HTML string and return the result. */
  mergeInto(html: string): string {
    if (this.removed.size === 0) return html;

    const doc = new DOMParser().parseFromString(html, 'text/html');
    let merged = 0;

    for (const [selector, collected] of this.removed) {
      if (collected.length === 0) continue;
      const parent = doc.querySelector(selector);
      if (!parent) continue;

      // Build a count map of outerHTML already present in the parent
      const existingCounts = new Map<string, number>();
      for (const child of Array.from(parent.children)) {
        const h = child.outerHTML;
        existingCounts.set(h, (existingCounts.get(h) || 0) + 1);
      }

      // Track how many of each collected HTML we've already skipped (matched to existing)
      const skippedCounts = new Map<string, number>();

      for (const nodeHTML of collected) {
        const existCount = existingCounts.get(nodeHTML) || 0;
        const skippedSoFar = skippedCounts.get(nodeHTML) || 0;

        // Skip if this node is already present in the DOM and we haven't
        // accounted for all existing copies yet
        if (skippedSoFar < existCount) {
          skippedCounts.set(nodeHTML, skippedSoFar + 1);
          continue;
        }

        const tpl = doc.createElement('template');
        tpl.innerHTML = nodeHTML;
        const child = tpl.content.firstElementChild;
        if (child) {
          parent.appendChild(child);
          merged++;
        }
      }
    }

    if (merged > 0) {
      console.log(`[Wayback] Merged ${merged} removed nodes back into snapshot`);
    }
    return doc.documentElement.outerHTML;
  }

  clear(): void {
    this.removed.clear();
    this.totalSize = 0;
  }

  get collectedCount(): number {
    let n = 0;
    for (const [, arr] of this.removed) n += arr.length;
    return n;
  }

  /** Remove one matching entry from any parent's array. */
  private removeOneMatch(html: string): void {
    for (const [, arr] of this.removed) {
      const idx = arr.indexOf(html);
      if (idx !== -1) {
        arr.splice(idx, 1);
        this.totalSize -= html.length;
        return;
      }
    }
  }

  /** Build a CSS selector path for an element. */
  private selectorFor(el: Element): string {
    const parts: string[] = [];
    let cur: Element | null = el;
    while (cur && cur !== document.documentElement) {
      if (cur.id) {
        parts.unshift('#' + CSS.escape(cur.id));
        break;
      }
      const parent: Element | null = cur.parentElement;
      if (!parent) break;
      const idx = Array.from(parent.children).indexOf(cur) + 1;
      parts.unshift(`${cur.tagName.toLowerCase()}:nth-child(${idx})`);
      cur = parent;
    }
    return parts.join(' > ') || 'body';
  }
}
