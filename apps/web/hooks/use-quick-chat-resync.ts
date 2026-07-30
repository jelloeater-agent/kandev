"use client";

import { useEffect, useRef } from "react";
import { useAppStore } from "@/components/state-provider";
import { listQuickChatSessions } from "@/lib/api/domains/workspace-api";
import { getStoredQuickChatNames } from "@/lib/local-storage";
import { migrateStoredQuickChatNames } from "@/lib/quick-chat/rename";
import type { QuickChatSession } from "@/lib/state/slices/ui/types";

/**
 * Keeps this client's quick-chat tabs in step with the server's list.
 *
 * Live changes arrive over the WebSocket, but a client that was asleep or
 * disconnected (a backgrounded phone tab is the common case) misses those
 * events entirely and would otherwise keep showing tabs nobody else has —
 * and keep missing tabs everyone else has — until a full page reload.
 * Re-reading the list whenever the socket (re)connects closes that gap.
 */
export function useQuickChatResync(workspaceId: string | null): void {
  const connectionStatus = useAppStore((state) => state.connection.status);
  const syncQuickChatSessions = useAppStore((state) => state.syncQuickChatSessions);
  // Resync once per connection, not on every unrelated status re-render.
  const lastSyncedConnection = useRef<string | null>(null);

  useEffect(() => {
    if (connectionStatus !== "connected") {
      lastSyncedConnection.current = null;
      return;
    }
    if (!workspaceId || lastSyncedConnection.current === workspaceId) return;
    lastSyncedConnection.current = workspaceId;

    let cancelled = false;
    listQuickChatSessions(workspaceId)
      .then((response) => {
        if (cancelled) return;
        const sessions: QuickChatSession[] = response.sessions.map((session) => ({
          kind: session.kind,
          sessionId: session.session_id,
          taskId: session.task_id,
          workspaceId: session.workspace_id,
          name: session.name,
          agentProfileId: session.agent_profile_id,
        }));
        syncQuickChatSessions(workspaceId, sessions);
        // Renames made before names were stored server-side live only in this
        // browser; push them up once so they reach the user's other devices.
        void migrateStoredQuickChatNames(sessions, getStoredQuickChatNames());
      })
      .catch(() => {
        // A failed resync must not clear the user's tabs; retry on next connect.
        if (!cancelled) lastSyncedConnection.current = null;
      });

    return () => {
      cancelled = true;
    };
  }, [connectionStatus, workspaceId, syncQuickChatSessions]);
}
