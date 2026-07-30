export type FileTreeEditorTarget = {
  /** Path relative to the resolved worktree root, as the editors API expects. */
  filePath: string;
  worktreeId?: string;
};

type WorktreeLike = { id: string; path?: string };

function pathBasename(path: string): string {
  return path.split(/[\\/]/).filter(Boolean).pop() ?? "";
}

/**
 * Maps a Files-panel tree path onto the `{worktreeId, filePath}` pair the
 * editors API resolves against.
 *
 * Single-worktree sessions serve the tree from inside the worktree, so node
 * paths are already worktree-relative. Multi-worktree sessions serve it from
 * the task root, where the first segment is the worktree directory and has to
 * be stripped before the rest can be resolved against that worktree.
 */
export function resolveFileTreeEditorTarget(
  nodePath: string,
  worktrees: WorktreeLike[],
): FileTreeEditorTarget {
  if (worktrees.length < 2) return { filePath: nodePath };

  const separatorIndex = nodePath.indexOf("/");
  const head = separatorIndex === -1 ? nodePath : nodePath.slice(0, separatorIndex);
  const rest = separatorIndex === -1 ? "" : nodePath.slice(separatorIndex + 1);
  const match = worktrees.find((worktree) => worktree.path && pathBasename(worktree.path) === head);
  if (!match) return { filePath: nodePath };
  return { filePath: rest, worktreeId: match.id };
}
