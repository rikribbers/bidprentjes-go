import csv
import random
import os


def generate_scans_test_data():
    os.makedirs("data", exist_ok=True)
    filename = "data/scans.csv"
    guid = "f95c14d3-c912-4fd2-8ab4-ea9d484e5f7b"

    # Pick 2000 random unique IDs from 1 to 100000 to ensure a good mix for testing
    ids = random.sample(range(1, 100001), 100000)

    with open(filename, "w", newline="") as f:
        writer = csv.writer(f)
        for bid_id in sorted(ids):
            # Sometimes a record has multiple scans; let's add a second scan for the first 5
            writer.writerow([str(bid_id), guid])
            if bid_id in sorted(ids)[:5]:
                writer.writerow([str(bid_id), guid])

    print(
        f"Generated {filename} with entries for {len(ids)} random bidprentjes using GUID {guid}"
    )


if __name__ == "__main__":
    generate_scans_test_data()
