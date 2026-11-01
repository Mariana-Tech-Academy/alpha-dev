import { useCallback, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import { api } from "../api/foodhub";
import { useAuth } from "../auth/context";
import { useResource } from "../hooks/useResource";
import { Empty, ErrorMessage, FoodImage, Loading } from "../components/ui";
import { hasMenuPhoto } from "../utils/images";
import { money } from "../utils/format";
export function RestaurantMenu() {
  const { id = "" } = useParams();
  const { user } = useAuth();
  const navigate = useNavigate();
  const load = useCallback(
    (signal: AbortSignal) => api.restaurant(id, signal),
    [id],
  );
  const { data, error, loading, retry } = useResource(load);
  const [category, setCategory] = useState("All");
  const [adding, setAdding] = useState<number | null>(null);
  const [actionError, setActionError] = useState<unknown>(null);
  const [message, setMessage] = useState("");
  async function add(itemId: number, name: string) {
    if (!user) {
      navigate("/login", { state: { from: `/restaurants/${id}` } });
      return;
    }
    setAdding(itemId);
    setActionError(null);
    setMessage("");
    try {
      await api.addItem(itemId);
      setMessage(`${name} added to your cart.`);
    } catch (err) {
      setActionError(err);
    } finally {
      setAdding(null);
    }
  }
  if (loading) return <Loading label="Loading the menu…" />;
  if (error || !data) return <ErrorMessage error={error} retry={retry} />;
  const menuItems = data.menu_items.filter((item) =>
    hasMenuPhoto(item.image_url),
  );
  const categories = [
    "All",
    ...new Set(menuItems.map((item) => item.category).filter(Boolean)),
  ];
  const activeCategory = categories.includes(category) ? category : "All";
  return (
    <section className="section">
      <Link className="back-link" to="/">
        ← All restaurants
      </Link>
      <div className="restaurant-banner">
        <div>
          <span className="eyebrow">THE MENU</span>
          <h1>{data.name}</h1>
          <p>{data.description}</p>
          <p className="muted">
            ⌖ {data.address}
            {data.city ? `, ${data.city}` : ""}
          </p>
          {data.opening_time && (
            <p className="muted">
              Opening hours: {data.opening_time}–{data.closing_time}
            </p>
          )}
        </div>
        <FoodImage src={data.image_url} name={data.name} />
      </div>
      <div className="section-heading">
        <h2>Find your favourite</h2>
        <Link className="button secondary" to="/cart">
          View cart →
        </Link>
      </div>
      <p className="subtle-note">
        Your cart can contain items from one restaurant at a time.
      </p>
      <div className="chips" aria-label="Filter menu by category">
        {categories.map((value) => (
          <button
            key={value}
            className={activeCategory === value ? "chip selected" : "chip"}
            aria-pressed={activeCategory === value}
            onClick={() => setCategory(value)}
          >
            {value}
          </button>
        ))}
      </div>
      <ErrorMessage error={actionError} />
      {message && (
        <p className="success" role="status">
          {message} <Link to="/cart">View cart →</Link>
        </p>
      )}
      {!menuItems.length ? (
        <Empty title="Nothing on the menu just yet">
          <p>Please check back for available dishes.</p>
        </Empty>
      ) : (
        <div className="menu-grid">
          {menuItems
            .filter(
              (item) =>
                activeCategory === "All" || item.category === activeCategory,
            )
            .map((item) => (
              <article className="menu-card" key={item.id}>
                <div className="menu-info">
                  <span className="eyebrow">{item.category}</span>
                  <h3>{item.name}</h3>
                  <p>{item.description}</p>
                  <div className="menu-action">
                    <strong>{money(item.price)}</strong>
                    <button
                      disabled={
                        adding !== null || (!!user && user.role !== "customer")
                      }
                      aria-label={`Add ${item.name} to cart`}
                      onClick={() => add(item.id, item.name)}
                    >
                      {adding === item.id ? "Adding…" : "+ Add"}
                    </button>
                  </div>
                </div>
                <FoodImage src={item.image_url} name={item.name} />
              </article>
            ))}
        </div>
      )}
      {user && user.role !== "customer" && (
        <p className="notice">Please use a customer account to place orders.</p>
      )}
    </section>
  );
}
