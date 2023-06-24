# import
import chromadb
from chromadb import Settings

client = chromadb.Client(Settings(
    chroma_db_impl="duckdb+parquet",
    persist_directory="./documents"
))
collection = client.create_collection(name="doc_collection")
results = collection.query(
    query_texts=["How do exec into a running container?"],
    n_results=2
)
