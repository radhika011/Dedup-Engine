import React, { useState } from "react";
import "./Pagination.css";
import { FaChevronRight, FaChevronLeft } from "react-icons/fa";
import { CgChevronDoubleLeft, CgChevronDoubleRight } from "react-icons/cg";
function Pagination({
  statusPerPage,
  totalStatus,
  paginate,
  currentPage,
  setCurrentPage,
}) {
  const [pageNumberLimit] = useState(4);
  const [maxPageNumberLimit, setMaxPageNumberLimit] = useState(4);
  const [minPageNumberLimit, setMinPageNumberLimit] = useState(0);

  const pageNumbers = [];

  for (let i = 1; i <= Math.ceil(totalStatus / statusPerPage); i++) {
    pageNumbers.push(i);
  }

  const handleFirst = () => {
    setCurrentPage(1);
    setMaxPageNumberLimit(4);
    setMinPageNumberLimit(0);
  };

  const handlePrev = () => {
    setCurrentPage(currentPage - 1);
    if ((currentPage - 1) % pageNumberLimit == 0) {
      setMaxPageNumberLimit(maxPageNumberLimit - pageNumberLimit);
      setMinPageNumberLimit(minPageNumberLimit - pageNumberLimit);
    }
  };
  const handleNext = () => {
    setCurrentPage(currentPage + 1);
    if (currentPage + 1 > maxPageNumberLimit) {
      setMaxPageNumberLimit(maxPageNumberLimit + pageNumberLimit);
      setMinPageNumberLimit(minPageNumberLimit + pageNumberLimit);
    }
  };

  const handleLast = () => {
    setCurrentPage(pageNumbers.length);
    setMaxPageNumberLimit(pageNumbers.length);
    setMinPageNumberLimit(pageNumbers.length - 4);
  };
  return (
    <nav>
      <ul className="pagination">
        <li>
          <button onClick={handleFirst}>
            <CgChevronDoubleLeft
              style={{ fontSize: "1.6rem" }}
            ></CgChevronDoubleLeft>
          </button>
        </li>
        <li>
          <button
            onClick={handlePrev}
            disabled={currentPage == pageNumbers[0] ? true : false}
          >
            <FaChevronLeft style={{ fontSize: "0.9rem" }}></FaChevronLeft>
          </button>
        </li>
        {pageNumbers.map((number) => {
          if (number < maxPageNumberLimit + 1 && number > minPageNumberLimit) {
            return (
              <li
                key={number}
                className={currentPage == number ? "active" : null}
              >
                <a onClick={() => paginate(number)} className="page-link">
                  {number}
                </a>
              </li>
            );
          } else return null;
        })}
        <li>
          <button
            onClick={handleNext}
            disabled={
              currentPage == pageNumbers[pageNumbers.length - 1] ? true : false
            }
          >
            <FaChevronRight style={{ fontSize: "0.9rem" }}></FaChevronRight>
          </button>
        </li>
        <li>
          <button onClick={handleLast}>
            <CgChevronDoubleRight
              style={{ fontSize: "1.6rem" }}
            ></CgChevronDoubleRight>
          </button>
        </li>
      </ul>
    </nav>
  );
}

export default Pagination;
