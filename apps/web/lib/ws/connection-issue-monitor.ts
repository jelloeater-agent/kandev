import type { ConnectionIssueSeverity, ConnectionStatus } from "@/lib/types/connection";

const UNSTABLE_DELAY_MS = 3_000;
const LOST_DELAY_MS = 10_000;

export class ConnectionIssueMonitor {
  private disposed = false;
  private offline = false;
  private unstableTimer: ReturnType<typeof setTimeout> | null = null;
  private lostTimer: ReturnType<typeof setTimeout> | null = null;

  constructor(private onSeverityChange: (severity: ConnectionIssueSeverity) => void) {}

  onStatusChange(status: ConnectionStatus) {
    if (this.disposed) {
      return;
    }

    if (status === "connected") {
      this.reset();
      return;
    }
    if (this.offline) return;

    this.offline = true;
    this.unstableTimer = setTimeout(() => this.onSeverityChange("unstable"), UNSTABLE_DELAY_MS);
    this.lostTimer = setTimeout(() => this.onSeverityChange("lost"), LOST_DELAY_MS);
  }

  dispose() {
    this.disposed = true;
    this.clearTimers();
  }

  private reset() {
    const wasOffline = this.offline;
    this.offline = false;
    this.clearTimers();
    if (wasOffline) this.onSeverityChange("none");
  }

  private clearTimers() {
    if (this.unstableTimer) clearTimeout(this.unstableTimer);
    if (this.lostTimer) clearTimeout(this.lostTimer);
    this.unstableTimer = null;
    this.lostTimer = null;
  }
}
