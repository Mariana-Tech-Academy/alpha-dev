import { useCallback } from "react";
import type { FormEvent } from "react";
import { Link, useSearchParams } from "react-router-dom";
import { hasRestaurantPhoto } from "../utils/images";
import { api } from "../api/foodhub";
import { useResource } from "../hooks/useResource";
import { Empty, ErrorMessage, FoodImage, Loading } from "../components/ui";
export function Restaurants() {
  const [params, setParams] = useSearchParams();
  const query = params.get("q")?.trim() || "";
  const page = Math.max(1, Number(params.get("page")) || 1);
  const load = useCallback(
    (signal: AbortSignal) => api.restaurants(query, page, signal),
    [query, page],
  );
  const { data, error, loading, retry } = useResource(load);
  const restaurants = data?.filter((restaurant) =>
    hasRestaurantPhoto(restaurant.image_url),
  );
  function search(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    const q = String(new FormData(event.currentTarget).get("q") || "").trim();
    setParams(q ? { q } : {});
  }
  const turnPage = (next: number) =>
    setParams({ ...(query ? { q: query } : {}), page: String(next) });
  return (
    <>
      <section className="hero">
        <div>
          <span className="eyebrow">A LITTLE LOCAL. A LOT OF FLAVOUR.</span>
          <h1>
            Your next favourite
            <br />
            meal starts here<span>.</span>
          </h1>
          <p>
            Discover local kitchens and find something delicious.
            <br className="desktop-break" /> Good food, just a few clicks away.
          </p>
          <form className="search" onSubmit={search}>
            <span aria-hidden="true">⌕</span>
            <input
              key={query}
              name="q"
              aria-label="Search restaurants or locations"
              placeholder="Search restaurants or locations"
              defaultValue={query}
            />
            <button>
              Find food <span aria-hidden="true">→</span>
            </button>
          </form>
        </div>
        <div className="hero-art hero-photograph">
          <FoodImage
            src="/demo-images/photos/hero/sushi-platter.jpg"
            name="Colourful sushi platter"
          />
        </div>
      </section>
      <section className="section">
        <div className="section-heading">
          <div>
            <span className="eyebrow">FIND YOUR FOOD MOOD</span>
            <h2>
              {query ? `Results for “${query}”` : "Explore local restaurants"}
            </h2>
          </div>
          <span className="muted">Fresh choices. Local favourites.</span>
        </div>
        {loading ? (
          <Loading />
        ) : error ? (
          <ErrorMessage error={error} retry={retry} />
        ) : !restaurants?.length ? (
          <Empty
            title={query ? "No restaurants found" : "The kitchens are quiet"}
          >
            <p>
              {query
                ? "Try another restaurant name, city or location."
                : "No active restaurants are available right now. Check back soon."}
            </p>
            {query && (
              <Link className="button secondary" to="/">
                Show all restaurants
              </Link>
            )}
          </Empty>
        ) : (
          <div className="restaurant-grid">
            {restaurants.map((restaurant) => (
              <Link
                className="restaurant-card"
                key={restaurant.id}
                to={`/restaurants/${restaurant.id}`}
              >
                <FoodImage src={restaurant.image_url} name={restaurant.name} />
                <div className="card-body">
                  <div className="card-title">
                    <h3>{restaurant.name}</h3>
                    <span className="round-arrow" aria-hidden="true">
                      ↗
                    </span>
                  </div>
                  <p className="clamp">{restaurant.description}</p>
                  <div className="card-meta">
                    <span>⌖ {restaurant.city || restaurant.address}</span>
                    <span>View menu →</span>
                  </div>
                </div>
              </Link>
            ))}
          </div>
        )}
        {!loading && !error && (page > 1 || (data?.length ?? 0) >= 12) && (
          <nav className="pagination" aria-label="Restaurant pages">
            <button
              className="secondary"
              disabled={page <= 1}
              onClick={() => turnPage(page - 1)}
            >
              ← Previous
            </button>
            <span>Page {page}</span>
            <button
              className="secondary"
              disabled={(data?.length ?? 0) < 12}
              onClick={() => turnPage(page + 1)}
            >
              Next →
            </button>
          </nav>
        )}
      </section>
      <aside className="bottom-note">
        <span aria-hidden="true">✳</span>
        <div>
          <h3>From their kitchen to your table.</h3>
          <p>
            Explore the menu, pick your favourites, and let dinner come
            together.
          </p>
        </div>
        <Link to="/orders" className="text-button">
          Your orders →
        </Link>
      </aside>
    </>
  );
}
