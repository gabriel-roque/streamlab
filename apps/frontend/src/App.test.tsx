import { fireEvent, render, screen, waitFor } from "@testing-library/react";
import { vi } from "vitest";
import App from "./App";

vi.mock("./lib/api", () => ({
  listVideos: vi.fn().mockRejectedValue(new Error("offline")),
  getPlayback: vi.fn().mockRejectedValue(new Error("offline")),
  createPlaybackSession: vi.fn().mockRejectedValue(new Error("offline")),
  sendPlaybackEvent: vi.fn().mockResolvedValue(undefined),
  uploadVideo: vi.fn(),
}));

describe("StreamLab library", () => {
  it("renders local assets when the API is unavailable", async () => {
    render(<App />);
    expect(await screen.findByText("Big Buck Bunny")).toBeInTheDocument();
    expect(screen.getByText("local fallback")).toBeInTheDocument();
    expect(screen.getByText("PROCESSING")).toBeInTheDocument();
  });

  it("filters assets by search and status", async () => {
    render(<App />);
    await screen.findByText("Big Buck Bunny");
    fireEvent.change(screen.getByLabelText("Search videos"), {
      target: { value: "camera" },
    });
    expect(screen.getByText("Camera ingest / studio B")).toBeInTheDocument();
    expect(screen.queryByText("Big Buck Bunny")).not.toBeInTheDocument();
    fireEvent.change(screen.getByLabelText("Filter by status"), {
      target: { value: "READY" },
    });
    expect(screen.getByText("No assets found")).toBeInTheDocument();
  });

  it("opens the upload dialog", async () => {
    render(<App />);
    await screen.findByText("Big Buck Bunny");
    fireEvent.click(screen.getByRole("button", { name: "Upload video" }));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Start ingest")).toBeInTheDocument();
  });

  it("opens a ready asset detail view", async () => {
    render(<App />);
    await screen.findByText("Big Buck Bunny");
    fireEvent.click(
      screen.getByRole("button", { name: "Open Big Buck Bunny" }),
    );
    await waitFor(() =>
      expect(screen.getByText("Asset metadata")).toBeInTheDocument(),
    );
    expect(screen.getByText("Playback observability")).toBeInTheDocument();
  });
});
