import { render, screen, waitFor } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import LoginPage from "./page";

const push = jest.fn();
const login = jest.fn();

jest.mock("next/navigation", () => ({
  useRouter: () => ({ push }),
  useSearchParams: () => new URLSearchParams(),
}));

jest.mock("@/context/AuthContext", () => ({
  useAuth: () => ({ login, register: jest.fn(), logout: jest.fn(), user: null, loading: false, refresh: jest.fn() }),
}));

describe("LoginPage", () => {
  beforeEach(() => {
    push.mockClear();
    login.mockClear();
  });

  it("shows validation errors for an empty submission", async () => {
    const user = userEvent.setup();
    render(<LoginPage />);

    await user.click(screen.getByRole("button", { name: "Log in" }));

    expect(await screen.findByText("Enter a valid email address")).toBeInTheDocument();
    expect(login).not.toHaveBeenCalled();
  });

  it("submits credentials and redirects on success", async () => {
    login.mockResolvedValueOnce(undefined);
    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(screen.getByLabelText("Email"), "alice@test.com");
    await user.type(screen.getByLabelText("Password"), "password1");
    await user.click(screen.getByRole("button", { name: "Log in" }));

    await waitFor(() => expect(login).toHaveBeenCalledWith("alice@test.com", "password1"));
    await waitFor(() => expect(push).toHaveBeenCalledWith("/"));
  });

  it("shows a server error message when login fails", async () => {
    login.mockRejectedValueOnce(new Error("invalid email or password"));
    const user = userEvent.setup();
    render(<LoginPage />);

    await user.type(screen.getByLabelText("Email"), "alice@test.com");
    await user.type(screen.getByLabelText("Password"), "wrong");
    await user.click(screen.getByRole("button", { name: "Log in" }));

    expect(await screen.findByText("invalid email or password")).toBeInTheDocument();
  });
});
