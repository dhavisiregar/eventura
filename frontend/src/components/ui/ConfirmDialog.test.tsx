import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useConfirmDialog } from "./ConfirmDialog";

function TestHarness({ onResult }: { onResult: (result: boolean) => void }) {
  const { confirm, dialog } = useConfirmDialog();
  return (
    <div>
      <button
        onClick={async () => {
          const result = await confirm({ title: "Delete this event?", description: "This cannot be undone.", danger: true });
          onResult(result);
        }}
      >
        Delete
      </button>
      {dialog}
    </div>
  );
}

describe("useConfirmDialog", () => {
  it("does not show the dialog until confirm() is called", () => {
    render(<TestHarness onResult={jest.fn()} />);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("resolves true when the confirm button is clicked", async () => {
    const user = userEvent.setup();
    const onResult = jest.fn();
    render(<TestHarness onResult={onResult} />);

    await user.click(screen.getByText("Delete"));
    expect(screen.getByRole("dialog")).toBeInTheDocument();
    expect(screen.getByText("Delete this event?")).toBeInTheDocument();

    await user.click(screen.getByText("Confirm"));
    expect(onResult).toHaveBeenCalledWith(true);
    expect(screen.queryByRole("dialog")).not.toBeInTheDocument();
  });

  it("resolves false when cancel is clicked", async () => {
    const user = userEvent.setup();
    const onResult = jest.fn();
    render(<TestHarness onResult={onResult} />);

    await user.click(screen.getByText("Delete"));
    await user.click(screen.getByText("Cancel"));

    expect(onResult).toHaveBeenCalledWith(false);
  });
});
