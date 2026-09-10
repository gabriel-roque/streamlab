import { fireEvent, render, screen } from "@testing-library/react";
import { vi } from "vitest";
import { Player } from "./Player";

vi.mock("hls.js", () => ({ default: class MockHls { static isSupported() { return false; } destroy() {} } }));

describe("Player", () => {
  it("renders accessible controls and quality selection", async () => {
    render(<Player source="https://example.com/video.mp4" protocol="MP4" title="Test source" />);
    const video = screen.getByLabelText("Player: Test source");
    expect(video).toBeInTheDocument();
    expect(video).toHaveProperty("muted", true);
    expect(screen.getByLabelText("Qualidade")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Reproduzir Test source" }));
    expect(await screen.findByRole("button", { name: "Pausar" })).toBeInTheDocument();
  });
});
