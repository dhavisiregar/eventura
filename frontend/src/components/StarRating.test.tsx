import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { StarRating } from "./StarRating";

describe("StarRating", () => {
  it("calls onChange with the clicked star value", async () => {
    const user = userEvent.setup();
    const onChange = jest.fn();
    render(<StarRating value={0} onChange={onChange} />);

    await user.click(screen.getByLabelText("4 stars"));
    expect(onChange).toHaveBeenCalledWith(4);
  });

  it("does not respond to clicks when readOnly", async () => {
    const user = userEvent.setup();
    const onChange = jest.fn();
    render(<StarRating value={3} onChange={onChange} readOnly />);

    await user.click(screen.getByLabelText("5 stars"));
    expect(onChange).not.toHaveBeenCalled();
  });
});
