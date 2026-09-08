import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { SearchBar } from "./SearchBar";

describe("SearchBar", () => {
  it("renders the current value", () => {
    render(<SearchBar value="jazz" onChange={jest.fn()} />);
    expect(screen.getByRole("searchbox")).toHaveValue("jazz");
  });

  it("calls onChange with the new value as the user types", async () => {
    const user = userEvent.setup();
    const onChange = jest.fn();
    render(<SearchBar value="" onChange={onChange} />);

    await user.type(screen.getByRole("searchbox"), "go");

    expect(onChange).toHaveBeenCalledWith("g");
    expect(onChange).toHaveBeenCalledWith("o");
  });
});
