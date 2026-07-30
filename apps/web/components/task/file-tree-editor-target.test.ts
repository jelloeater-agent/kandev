import { describe, expect, it } from "vitest";
import { resolveFileTreeEditorTarget } from "./file-tree-editor-target";

const PRIMARY = { id: "wt-1", path: "/tmp/task/kandev" };
const SIBLING = { id: "wt-2", path: "/tmp/task/kandev-feature-x" };
const README_PATH = "apps/web/README.md";

describe("resolveFileTreeEditorTarget", () => {
  it("passes the node path through for a session without worktrees", () => {
    expect(resolveFileTreeEditorTarget(README_PATH, [])).toEqual({
      filePath: README_PATH,
    });
  });

  it("passes the node path through for a single-worktree session", () => {
    expect(resolveFileTreeEditorTarget(README_PATH, [PRIMARY])).toEqual({
      filePath: README_PATH,
    });
  });

  it("strips the worktree directory segment for a multi-worktree session", () => {
    expect(
      resolveFileTreeEditorTarget(`kandev-feature-x/${README_PATH}`, [PRIMARY, SIBLING]),
    ).toEqual({ filePath: README_PATH, worktreeId: "wt-2" });
  });

  it("resolves the worktree directory itself to that worktree root", () => {
    expect(resolveFileTreeEditorTarget("kandev", [PRIMARY, SIBLING])).toEqual({
      filePath: "",
      worktreeId: "wt-1",
    });
  });

  it("falls back to the raw path when no worktree directory matches", () => {
    expect(resolveFileTreeEditorTarget("scratch/notes.md", [PRIMARY, SIBLING])).toEqual({
      filePath: "scratch/notes.md",
    });
  });

  it("matches worktree directories recorded with Windows separators", () => {
    const windowsSibling = { id: "wt-3", path: "C:\\tasks\\demo\\kandev-fix" };
    expect(
      resolveFileTreeEditorTarget(`kandev-fix/${README_PATH}`, [PRIMARY, windowsSibling]),
    ).toEqual({ filePath: README_PATH, worktreeId: "wt-3" });
  });

  it("ignores worktrees with no recorded path", () => {
    expect(resolveFileTreeEditorTarget("kandev/apps", [{ id: "wt-9" }, SIBLING])).toEqual({
      filePath: "kandev/apps",
    });
  });
});
