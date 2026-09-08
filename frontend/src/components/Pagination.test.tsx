import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Pagination } from "./Pagination";

describe("Pagination", () => {
  it("renders nothing when there is only one page", () => {
    const { container } = render(<Pagination page={1} totalPages={1} onChange={jest.fn()} />);
    expect(container).toBeEmptyDOMElement();
  });

  it("disables the previous button on the first page", () => {
    render(<Pagination page={1} totalPages={5} onChange={jest.fn()} />);
    expect(screen.getByLabelText("Previous page")).toBeDisabled();
    expect(screen.getByLabelText("Next page")).not.toBeDisabled();
  });

  it("disables the next button on the last page", () => {
    render(<Pagination page={5} totalPages={5} onChange={jest.fn()} />);
    expect(screen.getByLabelText("Next page")).toBeDisabled();
  });

  it("calls onChange with the target page when a page number is clicked", async () => {
    const user = userEvent.setup();
    const onChange = jest.fn();
    render(<Pagination page={1} totalPages={5} onChange={onChange} />);

    await user.click(screen.getByText("2"));
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it("always shows the last page even when it is far from the current page", async () => {
    const user = userEvent.setup();
    const onChange = jest.fn();
    render(<Pagination page={1} totalPages={5} onChange={onChange} />);

    await user.click(screen.getByText("5"));
    expect(onChange).toHaveBeenCalledWith(5);
  });

  it("calls onChange with page + 1 when next is clicked", async () => {
    const user = userEvent.setup();
    const onChange = jest.fn();
    render(<Pagination page={2} totalPages={5} onChange={onChange} />);

    await user.click(screen.getByLabelText("Next page"));
    expect(onChange).toHaveBeenCalledWith(3);
  });
});
