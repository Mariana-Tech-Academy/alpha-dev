import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import { AuthProvider } from "./auth/AuthProvider";
import { CustomerRoute } from "./auth/CustomerRoute";
import { Layout } from "./components/Layout";
import { Empty } from "./components/ui";
import { Restaurants } from "./pages/Restaurants";
import { RestaurantMenu } from "./pages/RestaurantMenu";
import { Auth } from "./pages/Auth";
import { Cart } from "./pages/Cart";
import { Checkout } from "./pages/Checkout";
import { Confirmation, Orders } from "./pages/Orders";
import "./App.css";
export default function App() {
  return (
    <BrowserRouter>
      <AuthProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route index element={<Restaurants />} />
            <Route path="restaurants/:id" element={<RestaurantMenu />} />
            <Route path="login" element={<Auth key="login" />} />
            <Route path="register" element={<Auth key="register" register />} />
            <Route element={<CustomerRoute />}>
              <Route path="cart" element={<Cart />} />
              <Route path="checkout" element={<Checkout />} />
              <Route path="orders" element={<Orders />} />
              <Route path="order-confirmation" element={<Confirmation />} />
            </Route>
            <Route
              path="*"
              element={
                <Empty title="This table hasn’t been set">
                  <p>We couldn’t find that page.</p>
                  <Link className="button" to="/">
                    Back to FoodHub
                  </Link>
                </Empty>
              }
            />
          </Route>
        </Routes>
      </AuthProvider>
    </BrowserRouter>
  );
}
